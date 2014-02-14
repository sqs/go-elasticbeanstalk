package main

import (
	"flag"
	"fmt"
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

func main() {
	http.Handle("/", http.HandlerFunc(hello))

	log.Printf("Listening on %s...", *bindAddr)
	err := http.ListenAndServe(*bindAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var t0 = time.Now()

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Go!")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uptime:", time.Since(t0))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "User-Agent:", r.UserAgent())
}
