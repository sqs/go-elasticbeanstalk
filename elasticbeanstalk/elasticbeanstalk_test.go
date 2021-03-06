package elasticbeanstalk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// Testing scheme adapted from go-github.

var (
	// mux i the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the GitHub client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with an elasticbeanstalk.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// elasticbeanstalk client configured to use test server
	client = NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func writeJSON(w http.ResponseWriter, jsonStr string) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	io.WriteString(w, jsonStr)
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if want != r.Method {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func mustParseTime(t *testing.T, timeStr string) Time {
	tm, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		t.Fatal("time.Parse(time.RFC3339Nano, %q) returned error: %v", timeStr, err)
	}
	tm = tm.Round(time.Millisecond)
	return Time{tm}
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(t)
	}
	return string(b)
}

func normTime(t *Time) {
	*t = Time{t.Time.UTC().Round(time.Second)}
}
