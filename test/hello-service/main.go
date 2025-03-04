package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Define HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: method=%s path=%s remote_addr=%s\n",
			r.Method, r.URL.Path, r.RemoteAddr)

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Hello, World!\n")
	})

	// Handle API /api/hello and return a json object
	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: method=%s path=%s remote_addr=%s\n",
			r.Method, r.URL.Path, r.RemoteAddr)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello, World!"}`)
	})

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
