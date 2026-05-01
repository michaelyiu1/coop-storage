package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bfbarry/coop-storage/metadata-server/config"
)

func setupTestDB(t *testing.T) func() {
	t.Helper()
	// Use a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "badger-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Override the DB path
	config.DB_PATH = tempDir

	// Initialize the database
	InitDb()

	// Return cleanup function
	return func() {
		CloseDb()
		os.RemoveAll(tempDir)
	}
}

func createTestMetaObject(t *testing.T, id, owner, fileName, fileType string) {
	t.Helper()

	metaObject := MetaObject{
		ID:         id,
		Owner:      owner,
		FileName:   fileName,
		FileType:   fileType,
		DeleteFlag: false,
		Version:    0,
	}

	if err := metaObject.Create(); err != nil {
		t.Fatalf("Failed to create test meta object: %v", err)
	}
}

func TestReadAllMetaObjects_HappyPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple metadata objects for the same user
	testUser := "testuser1"
	createTestMetaObject(t, "obj1", testUser, "file1.txt", "text/plain")
	createTestMetaObject(t, "obj2", testUser, "file2.jpg", "image/jpeg")
	createTestMetaObject(t, "obj3", testUser, "file3.pdf", "application/pdf")

	// Create objects for a different user to ensure they're not included
	createTestMetaObject(t, "obj4", "otheruser", "file4.txt", "text/plain")

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Make request
	resp, err := http.Get(fmt.Sprintf("%s/read_all_meta?user=%s", srv.URL, testUser))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var metaObjects []MetaObject
	if err := json.NewDecoder(resp.Body).Decode(&metaObjects); err != nil {
		t.Fatal(err)
	}

	// Verify we got exactly 3 objects
	if len(metaObjects) != 3 {
		t.Errorf("Expected 3 objects for %s, got %d", testUser, len(metaObjects))
	}

	// Verify all objects belong to the correct user
	fileNames := make(map[string]bool)
	for _, obj := range metaObjects {
		if obj.Owner != testUser {
			t.Errorf("Expected owner %s, got %s", testUser, obj.Owner)
		}
		fileNames[obj.FileName] = true
	}

	// Verify expected filenames are present
	expectedFiles := []string{"file1.txt", "file2.jpg", "file3.pdf"}
	for _, fileName := range expectedFiles {
		if !fileNames[fileName] {
			t.Errorf("Expected file %s not found in response", fileName)
		}
	}
}

func TestReadAllMetaObjects_UserNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	mux := http.NewServeMux()
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(fmt.Sprintf("%s/read_all_meta?user=nonexistentuser", srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestReadAllMetaObjects_MissingUserParameter(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	mux := http.NewServeMux()
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(fmt.Sprintf("%s/read_all_meta", srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestReadAllMetaObjects_MethodNotAllowed(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	mux := http.NewServeMux()
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, fmt.Sprintf("%s/read_all_meta?user=testuser", srv.URL), nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("method %s: want 405, got %d", method, resp.StatusCode)
			}
		})
	}
}

func TestReadAllMetaObjects_EmptyUserObjects(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create a user with no objects by creating and then deleting all objects
	testUser := "emptyuser"
	// Note: An empty user index will result in 404, which is expected behavior

	mux := http.NewServeMux()
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(fmt.Sprintf("%s/read_all_meta?user=%s", srv.URL, testUser))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// User with no objects should return 404
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404 for user with no objects, got %d", resp.StatusCode)
	}
}

func TestReadAllMetaObjects_ResponseContentType(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	testUser := "testuser"
	createTestMetaObject(t, "obj1", testUser, "file1.txt", "text/plain")

	mux := http.NewServeMux()
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(fmt.Sprintf("%s/read_all_meta?user=%s", srv.URL, testUser))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("want application/json Content-Type, got %q", ct)
	}
}
