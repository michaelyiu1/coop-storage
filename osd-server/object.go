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

// Write [TODO:description]
// Write [TODO:description]
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	bodyString := string(bodyBytes)
	log.Printf("Hello %s", bodyString)
	defer resp.Body.Close()
	// create file if all is well
	destPath := filepath.Join(UPLOADDIR, id)
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Failed to create file on server") // StatusInternalServerError
	}
	defer dest.Close()
	// this is what preview should look like

	// find current width and height of image,
	//
	//
	imageTypes := map[string]bool{
		".jpeg": true,
		".jpg":  true,
		".png":  true,
		".gif":  true,
	}
	if imageTypes[metadata.FileType] {

		img,_ := imaging.Open(metadata.FileName + metadata.FileType)
		preview,_ := imaging.Resize(img, PREVIEW_MAX_WIDTH, PREVIEW_MAX_HEIGHT, imaging.Lanczos)

	}

	// TODO: copy image
	//

	if _, err := io.Copy(dest, preview); err != nil {
		return fmt.Errorf("Failed to save image")
	}
	if _, err := io.Copy(dest, *file); err != nil {
		return fmt.Errorf("Failed to save file") // StatusInternalServerError
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("metadata server returned status: %d", resp.StatusCode)
	}

	return nil
}
