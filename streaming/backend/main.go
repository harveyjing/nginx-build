package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// FileInfo represents a file's metadata
type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	Path         string    `json:"path"`
	LastModified time.Time `json:"lastModified"`
	IsDirectory  bool      `json:"isDirectory"`
}

func main() {
	r := gin.Default()

	// API routes first
	api := r.Group("/api")
	{
		// Health check endpoint
		api.GET("/health", healthCheck)

		// File operations
		api.GET("/files", listFiles)
		api.GET("/download", handleDownload)
	}

	// Serve frontend static files
	r.NoRoute(func(c *gin.Context) {
		// Try to serve static files from the frontend directory
		fileServer := http.FileServer(http.Dir("../frontend"))
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	fmt.Println("Server starting at http://localhost:8080")
	fmt.Println("- Frontend: http://localhost:8080")
	fmt.Println("- API: http://localhost:8080/api")
	r.Run(":8080")
}

// healthCheck handles the health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// listFiles handles the file listing endpoint
func listFiles(c *gin.Context) {
	// Get the directory path from query parameter, default to data root
	dirPath := c.Query("path")
	if dirPath == "" {
		dirPath = "."
	}

	// Ensure the path is relative to data directory
	dataDir := "./data"
	fullPath := filepath.Join(dataDir, dirPath)

	// Security check - ensure the path is within data directory
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to resolve data directory: %v", err),
		})
		return
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to resolve path: %v", err),
		})
		return
	}

	if !filepath.HasPrefix(absFullPath, absDataDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid directory path",
		})
		return
	}

	// Check if directory exists
	dirInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Directory not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to access directory: %v", err),
			})
		}
		return
	}

	// Make sure it's a directory
	if !dirInfo.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path is not a directory",
		})
		return
	}

	// Read directory entries (non-recursive)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list directory contents: %v", err),
		})
		return
	}

	// Convert entries to FileInfo struct
	files := []FileInfo{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries with errors
		}

		entryPath := filepath.Join(dirPath, entry.Name())
		relPath, err := filepath.Rel(".", entryPath)
		if err != nil {
			relPath = entryPath // Fallback if relative path can't be determined
		}

		files = append(files, FileInfo{
			Name:         info.Name(),
			Size:         info.Size(),
			Path:         relPath,
			LastModified: info.ModTime(),
			IsDirectory:  info.IsDir(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"currentPath": dirPath,
		"files":       files,
	})
}

// handleDownload handles single and multiple file downloads
func handleDownload(c *gin.Context) {
	filePaths := c.QueryArray("files")
	if len(filePaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No files specified for download",
		})
		return
	}

	// Validate all file paths
	for _, path := range filePaths {
		fullPath := filepath.Join("./data", path)
		if !isPathSafe(fullPath) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid file path: %s", path),
			})
			return
		}
	}

	if len(filePaths) == 1 {
		handleSingleFileDownload(c, filePaths[0])
		return
	}

	handleMultiFileDownload(c, filePaths)
}

// handleSingleFileDownload streams a single file to the client
func handleSingleFileDownload(c *gin.Context, filePath string) {
	fullPath := filepath.Join("./data", filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to open file: %v", err),
		})
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get file info: %v", err),
		})
		return
	}

	// Set response headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Accept-Ranges", "bytes")

	// Stream the file
	http.ServeContent(c.Writer, c.Request, filepath.Base(filePath), fileInfo.ModTime(), file)
}

// handleMultiFileDownload creates a zip archive of multiple files and streams it
func handleMultiFileDownload(c *gin.Context, filePaths []string) {
	// Set response headers for chunked transfer
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=download.zip")
	c.Header("Transfer-Encoding", "chunked")

	// Create ZIP writer directly to response
	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	// Use larger buffer for better throughput
	buffer := make([]byte, 1024*1024) // 1MB buffer

	// Process each file
	for _, path := range filePaths {
		fullPath := filepath.Join("./data", path)

		// Get file info
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		// Open file
		file, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		// Create ZIP header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			file.Close()
			continue
		}

		header.Name = filepath.Base(path)
		header.Method = zip.Store

		// Create entry in ZIP
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			file.Close()
			continue
		}

		// Copy file data to ZIP
		_, err = io.CopyBuffer(writer, file, buffer)
		file.Close()
		if err != nil {
			continue
		}

		// Flush after each file to ensure data is sent
		c.Writer.Flush()
	}
}

// isPathSafe checks if the given file path is within the data directory
func isPathSafe(path string) bool {
	dataDir, err := filepath.Abs("./data")
	if err != nil {
		return false
	}

	filePath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	return filepath.HasPrefix(filePath, dataDir)
}
