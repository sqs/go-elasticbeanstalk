package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/kr/s3"
	"github.com/kr/s3/s3util"
	"github.com/sqs/go-elasticbeanstalk/elasticbeanstalk"
)

var verbose = flag.Bool("v", false, "show verbose output")

var elasticbeanstalkURL *url.URL

func initEnv() {
	var err error
	elasticbeanstalkURL, err = url.Parse(os.Getenv("ELASTICBEANSTALK_URL"))
	if err != nil {
		log.Fatal("Parsing ELASTICBEANSTALK_URL:", err)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ebc command [OPTS] ARGS...\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "The commands are:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tdeploy FILE...\t uploads and deploys a source bundle containing the specified files")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Environment variables:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tAWS_ACCESS_KEY_ID")
		fmt.Fprintln(os.Stderr, "\tAWS_SECRET_ACCESS_KEY")
		fmt.Fprintln(os.Stderr, "\tELASTICBEANSTALK_URL (default: https://elasticbeanstalk.us-east-1.amazonaws.com)")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Run `ebc command -h` for more information.")
		os.Exit(1)
	}

	flag.Parse()
	initEnv()

	if flag.NArg() == 0 {
		flag.Usage()
	}
	log.SetFlags(0)

	subcmd := flag.Arg(0)
	remaining := flag.Args()[1:]
	switch subcmd {
	case "deploy":
		deployCmd(remaining)
	}
}

func deployCmd(args []string) {
	fs := flag.NewFlagSet("deploy", flag.ExitOnError)
	env := fs.String("env", "", "EB environment name")
	app := fs.String("app", "", "EB application name")
	bucket := fs.String("bucket", "", "S3 bucket URL (example: https://example-bucket.s3-us-east-1.amazonaws.com)")
	label := fs.String("label", "", "label base name (suffix of -0, -1, -2, etc., is appended to ensure uniqueness)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ebc deploy [OPTS] FILE...\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Uploads and deploys a source bundle containing the specified files.")
		fmt.Fprintln(os.Stderr)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	fs.Parse(args)

	if *env == "" {
		fmt.Fprintln(os.Stderr, "env is required")
		fs.Usage()
	}

	if *app == "" {
		fmt.Fprintln(os.Stderr, "app is required")
		fs.Usage()
	}

	if *bucket == "" {
		fmt.Fprintln(os.Stderr, "bucket is required")
		fs.Usage()
	}
	bucketURL, err := url.Parse(*bucket)
	if err != nil {
		log.Fatal("parsing bucket URL:", err)
	}

	if *label == "" {
		fmt.Fprintln(os.Stderr, "label is required")
		fs.Usage()
	}

	err = deploy(fs.Args(), *env, *app, bucketURL, *label)
	if err != nil {
		log.Fatal("deploy:", err)
	}
}

var s3Config = s3util.Config{
	Keys: &s3.Keys{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	},
	Service: s3.DefaultService,
	Client:  http.DefaultClient,
}

func deploy(paths []string, env, app string, bucketURL *url.URL, label string) error {
	u, fullLabel, err := makeBundleObjectURL(bucketURL, label)
	if err != nil {
		return err
	}

	var t0 time.Time
	if *verbose {
		t0 = time.Now()
		log.Printf("Uploading source bundle to %s...", u.String())
	}
	w, err := s3util.Create(u.String(), nil, &s3Config)
	if err != nil {
		return err
	}
	err = writeZipArchive(paths, w)
	if err != nil {
		return err
	}
	if *verbose {
		log.Printf("Finished uploading source bundle (took %s).", time.Since(t0))
	}
	err = w.Close()
	if err != nil {
		return err
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		return err
	}
	c := elasticbeanstalk.Client{BaseURL: elasticbeanstalkURL, Auth: auth, Region: aws.Regions["us-west-2"]}

	if *verbose {
		log.Printf("Creating application version %q...", fullLabel)
	}
	err = c.CreateApplicationVersion(&elasticbeanstalk.CreateApplicationVersionParams{
		ApplicationName:      app,
		VersionLabel:         fullLabel,
		SourceBundleS3Bucket: "sg-eb-source-bundles",
		SourceBundleS3Key:    strings.TrimPrefix(u.Path, "/"),
	})
	if err != nil {
		return err
	}

	if *verbose {
		log.Printf("Updating environment %q to use version %q...", env, fullLabel)
	}
	err = c.UpdateEnvironment(&elasticbeanstalk.UpdateEnvironmentParams{
		EnvironmentName: env,
		VersionLabel:    fullLabel,
	})
	if err != nil {
		return err
	}

	return nil
}

// makeBundleObjectURL appends successive numeric prefixes to label until it
// finds a URL that doesn't refer to an existing object.
func makeBundleObjectURL(bucketURL *url.URL, label string) (*url.URL, string, error) {
	const max = 100
	for i := 0; i < max; i++ {
		fullLabel := fmt.Sprintf("%s-%d", label, i)
		u := s3URL(bucketURL, fullLabel+".zip")
		exists, err := s3ObjectExists(u.String())
		if err != nil {
			return nil, "", err
		}
		if !exists {
			return u, fullLabel, nil
		}
		if *verbose {
			log.Printf("Bundle exists at %s. Trying next suffix...", u.String())
		}
	}
	log.Fatal("bundles 0-%d with label %q already exist in bucket %s", max, label, bucketURL.String())
	panic("unreachable")
}

func s3URL(bucketURL *url.URL, key string) *url.URL {
	return bucketURL.ResolveReference(&url.URL{Path: key})
}

func s3ObjectExists(url string) (bool, error) {
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	s3Config.Sign(r, *s3Config.Keys)
	resp, err := s3Config.Client.Do(r)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected HTTP status code for %s: %d %s", url, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

func writeZipArchive(paths []string, w io.Writer) error {
	args := []string{"-r", "-", "--"}
	args = append(args, paths...)
	zip := exec.Command("zip", args...)
	zip.Stdout = w
	if *verbose {
		zip.Stderr = os.Stderr
	}
	return zip.Run()
}

func writeZipArchive_native(paths []string, w io.Writer) error {
	// DISABLED: seems to be incompatible with EB's zip file reading (yields
	// error: "Configuration files cannot be extracted from the application
	// version go-eb-38. Check that the application version is a valid zip or
	// war file.")

	// expand paths so that it lists all files.
	var filenames []string
	for _, path := range paths {
		path = filepath.Clean(path)
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			dirFiles, err := filesUnderDir(path)
			if err != nil {
				return err
			}
			if path != "." {
				filenames = append(filenames, path)
			}
			filenames = append(filenames, dirFiles...)
		} else {
			filenames = append(filenames, path)
		}
	}

	zw := zip.NewWriter(w)
	if *verbose {
		log.Printf("Writing %d files to source bundle...", len(filenames))
	}
	var totalBytes int64
	for _, filename := range filenames {
		if *verbose {
			log.Printf("- %s", filename)
		}
		fi, err := os.Stat(filename)
		if err != nil {
			return err
		}
		h := &zip.FileHeader{Name: filename}
		h.SetModTime(fi.ModTime())
		h.SetMode(fi.Mode())
		if fi.Mode().IsDir() {
			h.Name += "/"
		} else {
			h.Method = zip.Deflate
		}
		f, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if !fi.Mode().IsDir() {
			file, err := os.Open(filename)
			if err != nil {
				return err
			}
			n, err := io.Copy(f, file)
			if err != nil {
				return err
			}
			totalBytes += n
		}
	}
	err := zw.Close()
	if err != nil {
		return err
	}
	if *verbose {
		log.Printf("Finished creating source bundle archive (%d MB uncompressed)", totalBytes/1024/1024)
	}

	return nil
}

func filesUnderDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() || info.Mode().IsDir() && path != "." {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
