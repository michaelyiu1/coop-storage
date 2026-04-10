package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"os"
	"log"
	"encoding/json"
	"strings"
)

type OSDGuide struct {
	Token string `json:"Token"`
	PageNum int `json:"PageNum"`
	IDMap map[string]string `json:"IDMap"`
}

var (
	OSDSERVERBASE = "http://localhost:8280"
	METASERVERBASE = "http://localhost:7676"
	FILENAME = "testdocx.docx"
	TESTDATADIR = "/Users/ethanspraggon/Projects/coop-storage/tests/data"

	FILEPATH = fmt.Sprintf("%s/%s", TESTDATADIR, FILENAME)
	TESTUSER = "placeholder"
)

func main() {
	// upload
	if err := uploadFile(); err != nil {
		log.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Upload completed.")

	
	// download 
	if err := downloadFile(); err != nil {
		log.Printf("Download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Download completed.")

	// TODO: test update

	// TODO: test deletion

}

func uploadFile() error {

	if _, err := os.Stat(FILEPATH); os.IsNotExist(err) {
		log.Printf("Error: File '%s' does not exist\n", FILEPATH)
		os.Exit(1)
	}

	uploadEndpoint := fmt.Sprintf("%s/upload", OSDSERVERBASE)

	// Open the file
	file, err := os.Open(FILEPATH)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a buffer to write our multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file field
	part, err := writer.CreateFormFile("file", filepath.Base(FILEPATH))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file contents to the form field
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	objId, err := httpRequest("POST", uploadEndpoint, body, writer)
	if err != nil {
		return err
	}

	// //TODO: check if metadata server is populated properly

	res, err := httpRequest(
		"GET", 
		fmt.Sprintf("%s/read_meta?id=%s", METASERVERBASE, string(objId)),
		nil,
		nil,
	)

	log.Println(string(res))
	return nil
}

func downloadFile() error {
	// HERE: ping metadata for object ids
	res, err := httpRequest(
		"GET", 
		fmt.Sprintf("%s/prepare_osd_request?user=%s", METASERVERBASE, TESTUSER),
		nil,
		nil,
	)
	
	if err != nil {
		return err
	}

	g := OSDGuide{}
	json.Unmarshal(res, &g)
	var objId string
	for name, id := range g.IDMap {
		if name == FILENAME {
			objId = id
			break
		}
	}

	// HERE: get the data from OSD server
	fileBytes, err := httpRequest(
		"GET", 
		fmt.Sprintf("%s/download/%s", OSDSERVERBASE, objId),
		nil,
		nil,
	)

	if err != nil {
		return err
	}

	parts := strings.Split(FILENAME, ".")
	fileOut := fmt.Sprintf("%s_downloaded.%s", parts[0], parts[1])
	err = os.WriteFile(fmt.Sprintf("%s/%s", TESTDATADIR, fileOut), fileBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write bytes to file: %w", err)
	}

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
