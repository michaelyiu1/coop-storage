package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"log"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <server-url> <jpeg-file-path>")
		os.Exit(1)
	}

	serverBase := "http://localhost:8280"
	filePath := os.Args[1]
	serverURL := fmt.Sprintf("%s/upload", serverBase)

	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Error: File '%s' does not exist\n", filePath)
		os.Exit(1)
	}

	log.Printf("Uploading file: %s\n", filePath)
	
	if err := uploadFile(filePath, serverURL); err != nil {
		log.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Upload completed successfully!")
}

func uploadFile(filePath, serverURL string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a buffer to write our multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file contents to the form field
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Close the writer to finalize the multipart message
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", serverURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type with boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(responseBody))
	}

	log.Printf("Server response: %s", string(responseBody))
	return nil
}