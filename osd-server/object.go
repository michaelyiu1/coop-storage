package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type ObjectFile struct {
	// METADATA
	Id       string // just used for client
	Contents []byte
}

type MetadataPOST struct {
	ID       string `json:"id"`
	FileType string `json:"fileType"`
	FileName string `json:"fileName"`
}

// Write stores a file upload by assigning it a new UUID, registering its
// metadata with the metadata server, and saving the file contents (along with
// an image preview for supported image types) to the upload directory.
// It returns an error if any step fails.
func (o *ObjectFile) Write(file *multipart.File, header *multipart.FileHeader) error {
	// TODO: parallel writes
	id := uuid.New().String()
	o.Id = id
	// write to metadata server
	metadata := MetadataPOST{
		ID:       id,
		FileType: filepath.Ext(header.Filename),
		FileName: header.Filename,
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	endpoint := fmt.Sprintf("%s/write_meta", METADATASERVERURL)
	resp, perr := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	log.Printf("just sent a post to %s", endpoint)
	if perr != nil {
		return fmt.Errorf("failed to POST to metadata server: %w", perr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("metadata server returned status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("Hello %s", string(bodyBytes))

	imageTypes := map[string]bool{
		".jpeg": true,
		".jpg":  true,
		".png":  true,
		".gif":  true,
	}
	if imageTypes[metadata.FileType] {
		img, err := imaging.Decode(*file)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		preview := imaging.Resize(img, PREVIEW_MAX_WIDTH, PREVIEW_MAX_HEIGHT, imaging.Lanczos)
		previewPath := filepath.Join(UPLOADPVW, metadata.ID+".jpg")
		if err := imaging.Save(preview, previewPath); err != nil {
			return fmt.Errorf("failed to save preview: %w", err)
		}
		if _, err := (*file).Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
	}

	destPath := filepath.Join(UPLOADDIR, id)
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file at %s: %w", destPath, err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, *file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
