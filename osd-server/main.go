package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Create store directory if it doesn't exist
	if err := os.MkdirAll(UPLOADDIR, 0755); err != nil {
		log.Fatal("Failed to create store directory:", err)
	}

	http.Handle("/store/", http.StripPrefix("/store/", http.FileServer(http.Dir("./store"))))

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})
	http.HandleFunc("/preview/", previewHandler)
	log.Printf("Server starting on PORT %s\n", PORT)
	log.Printf("Upload endpoint: http://localhost%s/upload\n", PORT)
	log.Printf("Download endpoint: http://localhost%s/download/{filename}\n", PORT)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func previewHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	html := fmt.Sprintf(`<!DOCTYPE html>
	<html>
	<body>
		<img src="/store/%s.jpg" alt="Preview" style="max-width: 256px;">
	</body>
	</html>`, filename)

	fmt.Fprint(w, html)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, MAXUPLOADSIZE)
	
	if err := r.ParseMultipartForm(MAXUPLOADSIZE); err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	o := ObjectFile{}
	err = o.Write(&file, header)
	if err != nil {
		log.Printf("bad write %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	log.Printf("File uploaded successfully: %s\n", header.Filename)
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s\n", header.Filename)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from URL path
	filename := filepath.Base(r.URL.Path)
	if filename == "." || filename == "/" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(UPLOADDIR, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set headers
	// TODO switch for certain types to be able to preview e.g., image/jpeg
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Copy file to response
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("Error sending file: %v\n", err)
		return
	}

	log.Printf("File downloaded: %s\n", filename)
}