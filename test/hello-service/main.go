package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for testing
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

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

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Basic hello endpoint
	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: method=%s path=%s remote_addr=%s\n",
			r.Method, r.URL.Path, r.RemoteAddr)
		fmt.Fprintf(w, "Hello, World!\n")
	})

	// File endpoint handling both download (GET) and upload (POST)
	r.HandleFunc("/api/file/{uuid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["uuid"]

		// Validate UUID format
		if _, err := uuid.Parse(fileID); err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleFileDownload(w, r, fileID)
		case http.MethodPost:
			handleFileUpload(w, r, fileID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// WebSocket endpoint
	r.HandleFunc("/api/ws", handleWebSocket)

	// WebSocket test page
	r.HandleFunc("/ws-test", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/ws-test.html")
	})

	// JSON endporing
	r.HandleFunc("/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello, World!"}`)
	})

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Send welcome message
	welcome := Message{
		Type:    "welcome",
		Content: "Connected to WebSocket server",
	}
	if err := conn.WriteJSON(welcome); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Message handling loop
	for {
		// Read message
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Log received message
		log.Printf("Received message from %s: %+v", r.RemoteAddr, msg)

		// Echo the message back
		response := Message{
			Type:    "echo",
			Content: msg.Content,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Failed to send echo message: %v", err)
			break
		}
	}
}

func handleFileDownload(w http.ResponseWriter, r *http.Request, fileID string) {
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
}

func handleFileUpload(w http.ResponseWriter, r *http.Request, fileID string) {
	// Check content length
	contentLength := r.ContentLength
	if contentLength <= 0 {
		http.Error(w, "Content-Length required", http.StatusBadRequest)
		return
	}

	// Read the uploaded data
	var totalBytes int64
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := r.Body.Read(buffer)
		if n > 0 {
			totalBytes += int64(n)
			// In a real implementation, we would write this data somewhere
			// For now, we just count the bytes
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading upload: %v", err)
			http.Error(w, "Error reading upload", http.StatusInternalServerError)
			return
		}
	}

	// Log the upload
	log.Printf("Received file upload: uuid=%s size=%d remote_addr=%s\n",
		fileID, totalBytes, r.RemoteAddr)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"uuid":    fileID,
		"size":    totalBytes,
	})
}
