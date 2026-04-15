package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	// "strings"
)

type OSDGuide struct {
	Token   string            `json:"Token"`
	PageNum int               `json:"PageNum"`
	IDMap   map[string]string `json:"IDMap"`
}

var (
	METASERVERBASE = "http://localhost:7678"
	FILENAME       = "ball.jpg"
	TESTDATADIR    = "/Users/Michael/Documents/test_images"

	FILEPATH = fmt.Sprintf("%s/%s", TESTDATADIR, FILENAME)
	TESTUSER = "placeholder"
)

func main() {
	objectKey, err := uploadFile()
	if err != nil {
		log.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Upload completed.")

	if err := downloadFile(objectKey); err != nil {
		log.Printf("Download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Download completed.")

	// TODO: test update
	// TODO: test deletion
}

func uploadFile() (string, error) {
	if _, err := os.Stat(FILEPATH); os.IsNotExist(err) {
		return "", fmt.Errorf("file '%s' does not exist", FILEPATH)

	}

	// Step 1: get presigned upload URL from metadata server
	fileInfo, err := os.Stat(FILEPATH)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	presignBody, _ := json.Marshal(map[string]any{
		"filename":       FILENAME,
		"content_type":   "application/octet-stream",
		"content_length": fileInfo.Size(),
	})

	resp, err := http.Post(
		METASERVERBASE+"/upload/presign",
		"application/json",
		bytes.NewReader(presignBody),
	)
	if err != nil {
		return "", fmt.Errorf("presign request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("presign returned %d: %s", resp.StatusCode, body)
	}

	var presign struct {
		UploadURL string `json:"upload_url"`
		ObjectKey string `json:"object_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&presign); err != nil {
		return "", fmt.Errorf("failed to decode presign response: %w", err)
	}
	log.Printf("Got presign URL for object_key: %s", presign.ObjectKey)

	// Step 2: PUT file bytes directly to RustFS presigned URL
	file, err := os.Open(FILEPATH)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	putReq, err := http.NewRequest(http.MethodPut, presign.UploadURL, file)
	if err != nil {
		return "", fmt.Errorf("failed to create PUT request: %w", err)
	}
	putReq.Header.Set("Content-Type", "application/octet-stream")
	putReq.ContentLength = fileInfo.Size()

	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return "", fmt.Errorf("PUT to RustFS failed: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(putResp.Body)
		return "", fmt.Errorf("RustFS PUT returned %d: %s", putResp.StatusCode, body)
	}

	log.Printf("File uploaded successfully. ObjectKey: %s", presign.ObjectKey)
	return presign.ObjectKey, nil
}

func downloadFile(objectKey string) error {
	resp, err := http.Get(
		fmt.Sprintf("%s/download/presign/%s", METASERVERBASE, url.PathEscape(objectKey)),
	)
	if err != nil {
		return fmt.Errorf("download presign request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download presign returned %d: %s", resp.StatusCode, body)
	}

	var presign struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&presign); err != nil {
		return fmt.Errorf("failed to decode presign response: %w", err)
	}
	log.Printf("Got download URL: %s", presign.DownloadURL)

	getResp, err := http.Get(presign.DownloadURL)
	if err != nil {
		return fmt.Errorf("GET from RustFS failed: %w", err)
	}
	defer getResp.Body.Close()

	outPath := fmt.Sprintf("%s/downloaded_%s", TESTDATADIR, FILENAME)
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, getResp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("File downloaded to %s", outPath)
	return nil
}

////////////
// UTILS  //
////////////

func httpRequest(mode string, url string, body *bytes.Buffer, writer *multipart.Writer) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = body
	}
	req, err := http.NewRequest(mode, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type with boundary
	if writer != nil {
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
