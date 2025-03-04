package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RandomDataReader generates random data of specified size
type RandomDataReader struct {
	size   int64
	offset int64
}

func (r *RandomDataReader) Read(p []byte) (n int, err error) {
	if r.offset >= r.size {
		return 0, io.EOF
	}

	remaining := r.size - r.offset
	toRead := int64(len(p))
	if toRead > remaining {
		toRead = remaining
	}

	for i := int64(0); i < toRead; i++ {
		p[i] = byte(rand.Intn(256))
	}

	r.offset += toRead
	return int(toRead), nil
}

func (r *RandomDataReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.offset + offset
	case io.SeekEnd:
		abs = r.size + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if abs < 0 {
		return 0, fmt.Errorf("negative position: %d", abs)
	}

	r.offset = abs
	return abs, nil
}

func main() {
	// Initialize router
	r := mux.NewRouter()

	// Basic hello endpoint
	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: method=%s path=%s remote_addr=%s\n",
			r.Method, r.URL.Path, r.RemoteAddr)
		fmt.Fprintf(w, "Hello, World!\n")
	})

	// File download endpoint with random data
	r.HandleFunc("/api/file/{uuid}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vars := mux.Vars(r)
		fileID := vars["uuid"]

		// Validate UUID format
		if _, err := uuid.Parse(fileID); err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		// Use query parameter for file size, default to 1MB
		size := int64(1024 * 1024) // 1MB default
		if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
			if parsedSize, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				size = parsedSize
			}
		}

		// Set response headers
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=random-%s.bin", fileID))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

		// Log the download request
		log.Printf("Serving random file: uuid=%s size=%d remote_addr=%s\n",
			fileID, size, r.RemoteAddr)

		// Create random data reader and serve content
		reader := &RandomDataReader{size: size}
		http.ServeContent(w, r, fileID, time.Now(), reader)
	})

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
