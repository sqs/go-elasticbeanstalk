package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var bindAddr = flag.String("http", ":8080", "http listen address")

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
