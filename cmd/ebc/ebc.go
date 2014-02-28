package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/jteeuwen/ini"
	"github.com/kr/s3"
	"github.com/kr/s3/s3util"
	"github.com/sqs/go-elasticbeanstalk/elasticbeanstalk"
)

var dir = flag.String("dir", ".", "dir to operate in")
var verbose = flag.Bool("v", false, "show verbose output")
var debugKeepTempDirs = flag.Bool("debug.keep-temp-dirs", false, "(debug) don't remove temp dirs")

var elasticbeanstalkURL *url.URL
var ebClient *elasticbeanstalk.Client

var t0 = time.Now()

func initEnv() {
	var err error
	elasticbeanstalkURL, err = url.Parse(os.Getenv("ELASTICBEANSTALK_URL"))
	if err != nil {
		log.Fatal("Parsing ELASTICBEANSTALK_URL:", err)
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}
	ebClient = &elasticbeanstalk.Client{BaseURL: elasticbeanstalkURL, Auth: auth, Region: aws.Regions["us-west-2"]}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ebc command [OPTS] ARGS...\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "The commands are:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tbundle\t creates a source bundle for a directory (running scripts if they exist)")
		fmt.Fprintln(os.Stderr, "\tdeploy\t deploys a directory")
		fmt.Fprintln(os.Stderr, "\tupload BUNDLE-FILE\t uploads the source bundle")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Environment variables:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tAWS_ACCESS_KEY_ID")
		fmt.Fprintln(os.Stderr, "\tAWS_SECRET_KEY")
		fmt.Fprintln(os.Stderr, "\tELASTICBEANSTALK_URL (default: https://elasticbeanstalk.us-east-1.amazonaws.com)")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Run `ebc command -h` for more information.")
		os.Exit(1)
	}

	flag.Parse()
	initEnv()

	var err error
	*dir, err = filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	if flag.NArg() == 0 {
		flag.Usage()
	}
	log.SetFlags(0)

	subcmd := flag.Arg(0)
	remaining := flag.Args()[1:]
	switch subcmd {
	case "bundle":
		bundleCmd(remaining)
	case "deploy":
		deployCmd(remaining)
	case "upload":
		uploadCmd(remaining)
	}
}

const bundleScript = ".ebc-bundle"

func bundleCmd(args []string) {
	fs := flag.NewFlagSet("bundle", flag.ExitOnError)
	outFile := fs.String("out", "eb-bundle.zip", "output file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ebc bundle\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Creates a source bundle for a directory (specified with -dir=DIR). If the directory contains an %s file, it is executed with a temporary output directory as its first argument, and it's expected to write the source bundle to that directory. Otherwise, if no %s file exists, the directory itself is used as the bundle source.", bundleScript, bundleScript)
		fmt.Fprintln(os.Stderr)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	fs.Parse(args)

	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "no positional args")
		fs.Usage()
	}

	fw, err := os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()
	err = bundle(*dir, fw)
	if err != nil {
		log.Fatal("bundle failed: ", err)
	}

	fi, err := os.Stat(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote bundle file: %s (%.1f MB, took %s)\n", *outFile, float64(fi.Size())/1024/1024, time.Since(t0))
}

func bundle(dir string, w io.Writer) error {
	scriptFile := filepath.Join(dir, bundleScript)
	fi, err := os.Stat(scriptFile)
	if err == nil && fi.Mode().IsRegular() {
		if *verbose {
			log.Printf("Running bundle script %s...", scriptFile)
		}
		tmpDir, err := ioutil.TempDir("", "ebc")
		if err != nil {
			return err
		}
		if *debugKeepTempDirs {
			log.Printf("Writing bundle output to temp dir %s", tmpDir)
		} else {
			defer os.RemoveAll(tmpDir)
		}
		script := exec.Command(scriptFile, tmpDir)
		script.Dir = dir
		if *verbose {
			script.Stdout, script.Stderr = os.Stderr, os.Stderr
		}
		err = script.Run()
		if err != nil {
			return fmt.Errorf("running %s: %s", scriptFile, err)
		}
		dir = tmpDir
	}

	return writeZipArchive(dir, w)
}

type defaults struct {
	env       string
	app       string
	bucketURL string
	label     string
}

func readDefaults(dir string) (*defaults, error) {
	currentBranch, err := cmdOutput(dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(dir, ".elasticbeanstalk/config")
	ini := ini.New()
	err = ini.Load(configFile)
	if err != nil {
		return nil, err
	}

	d := new(defaults)

	branchConfig := ini.Section("branch:" + currentBranch)
	globalConfig := ini.Section("global")
	var get = func(key string, default_ string) string {
		if branchConfig != nil {
			bv := branchConfig.S(key, default_)
			if bv != "" {
				return bv
			}
		}
		return globalConfig.S(key, default_)
	}

	d.app = get("ApplicationName", "")
	d.env = get("EnvironmentName", "")
	d.label = filepath.Base(dir)
	region := get("Region", "")
	if d.app != "" && region != "" {
		d.bucketURL = fmt.Sprintf("https://eb-bundle-%s.s3-%s.amazonaws.com", d.app, region)
	}

	if *verbose {
		log.Printf("Read defaults for branch %q from %s: %+v", currentBranch, configFile, d)
	}
	return d, nil
}

func cmdOutput(cwd, exe string, args ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	cmd.Dir = cwd
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func uploadCmd(args []string) {
	df, err := readDefaults(*dir)
	if err != nil && *verbose {
		log.Printf("Warning: couldn't read defaults: %s", err)
		df = new(defaults)
	}

	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	env := fs.String("env", df.env, "EB environment name")
	app := fs.String("app", df.app, "EB application name")
	bucket := fs.String("bucket", df.bucketURL, "S3 bucket URL (example: https://example-bucket.s3-us-east-1.amazonaws.com)")
	label := fs.String("label", df.label, "label base name (suffix of -0, -1, -2, etc., is appended to ensure uniqueness)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ebc upload [OPTS] BUNDLE-FILE\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Uploads the specified source bundle.")
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

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "exactly 1 bundle file must be specified")
		fs.Usage()
	}
	bundleFile := fs.Arg(0)
	f, err := os.Open(bundleFile)
	if err != nil {
		log.Fatal(err)
	}

	fullLabel, err := upload(f, *env, *app, bucketURL, *label)
	if err != nil {
		log.Fatal("upload failed: ", err)
	}
	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Uploaded %s as label %q (%.1f MB, took %s)\n", bundleFile, fullLabel, float64(fi.Size())/1024/1024, time.Since(t0))
}

var s3Config = s3util.Config{
	Keys: &s3.Keys{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_KEY"),
	},
	Service: s3.DefaultService,
	Client:  http.DefaultClient,
}

func upload(r io.Reader, env, app string, bucketURL *url.URL, label string) (string, error) {
	u, fullLabel, err := makeBundleObjectURL(bucketURL, label)
	if err != nil {
		return "", err
	}

	if *verbose {
		log.Printf("Uploading source bundle to %s...", u.String())
	}

	w, err := s3util.Create(u.String(), nil, &s3Config)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}

	if *verbose {
		log.Printf("Creating application version %q...", fullLabel)
	}
	err = ebClient.CreateApplicationVersion(&elasticbeanstalk.CreateApplicationVersionParams{
		ApplicationName:      app,
		VersionLabel:         fullLabel,
		SourceBundleS3Bucket: s3BucketFromURL(u),
		SourceBundleS3Key:    strings.TrimPrefix(u.Path, "/"),
	})
	if err != nil {
		return "", err
	}

	return fullLabel, nil
}

func s3BucketFromURL(u *url.URL) string {
	return strings.Split(u.Host, ".")[0]
}

func deployCmd(args []string) {
	df, err := readDefaults(*dir)
	if err != nil && *verbose {
		log.Printf("Warning: couldn't read defaults: %s", err)
		df = new(defaults)
	}

	fs := flag.NewFlagSet("deploy", flag.ExitOnError)
	env := fs.String("env", df.env, "EB environment name")
	app := fs.String("app", df.app, "EB application name")
	bucket := fs.String("bucket", df.bucketURL, "S3 bucket URL (example: https://example-bucket.s3-us-east-1.amazonaws.com)")
	label := fs.String("label", df.label, "label base name (suffix of -0, -1, -2, etc., is appended to ensure uniqueness)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ebc deploy [OPTS]\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Bundles and deploys a directory (specified with -dir=DIR).")
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

	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "no positional args")
		fs.Usage()
	}

	err = deploy(*dir, *env, *app, bucketURL, *label)
	if err != nil {
		log.Fatal("deploy failed: ", err)
	}
	fmt.Printf("Deploy initiated (took %s)\n", time.Since(t0))
}

func deploy(dir string, env, app string, bucketURL *url.URL, label string) error {
	var buf bytes.Buffer
	err := bundle(dir, &buf)
	if err != nil {
		return err
	}

	var fullLabel string
	fullLabel, err = upload(&buf, env, app, bucketURL, label)
	if err != nil {
		return err
	}

	if *verbose {
		log.Printf("Updating environment %q to use version %q...", env, fullLabel)
	}
	err = ebClient.UpdateEnvironment(&elasticbeanstalk.UpdateEnvironmentParams{
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

func writeZipArchive(dir string, w io.Writer) error {
	zip := exec.Command("zip", "-r", "-", ".")
	zip.Dir = dir
	zip.Stdout = w
	if *verbose {
		zip.Stderr = os.Stderr
	}
	err := zip.Run()
	if err != nil {
		return fmt.Errorf("writing zip archive: %s", err)
	}
	return nil
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
