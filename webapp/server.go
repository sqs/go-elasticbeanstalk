package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8888"
}

var bindAddr = flag.String("http", ":"+port(), "http listen address")
var gitCommitID, gitBranch string

func init() {
	data, err := ioutil.ReadFile("./.git-commit-id")
	if err != nil {
		gitCommitID = fmt.Sprintf("error: %s", err)
	}
	gitCommitID = string(data)

	data, err = ioutil.ReadFile("./.git-branch")
	if err != nil {
		gitBranch = fmt.Sprintf("error: %s", err)
	}
	gitBranch = string(data)
}

func main() {
	flag.Parse()

	http.Handle("/", http.HandlerFunc(hello))

	log.Printf("Listening on %s...", *bindAddr)
	err := http.ListenAndServe(*bindAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var t0 = time.Now()
var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal("Hostname:", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Go!")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Git commit ID:", gitCommitID)
	fmt.Fprintln(w, "Git branch:", gitBranch)
	fmt.Fprintln(w, "Uptime:", time.Since(t0))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "`hostname`:", hostname)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Request headers:")
	for k, vs := range r.Header {
		for _, v := range vs {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
	}
}
