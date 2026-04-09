package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	handler "github.com/bfbarry/coop-storage/metadata-server/controllers"
)

// mockUploader lets tests control what the storage layer returns.
type mockUploader struct {
	url       string
	expiresAt time.Time
	err       error
}

func (m *mockUploader) PresignUpload(_ context.Context, _, _ string, _ int64) (string, time.Time, error) {
	return m.url, m.expiresAt, m.err
}

func newTestServer(t *testing.T, uploader handler.Uploader) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handler.NewUploadHandler(uploader).Register(mux)
	return httptest.NewServer(mux)
}

func TestPresign_HappyPath(t *testing.T) {
	expiry := time.Now().Add(15 * time.Minute)
	mock := &mockUploader{url: "http://rustfs/presigned-put-url", expiresAt: expiry}
	srv := newTestServer(t, mock)
	defer srv.Close()

	body := `{"filename":"photo.jpg","content_type":"image/jpeg","content_length":1048576}`
	resp, err := http.Post(srv.URL+"/upload/presign", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var got handler.PresignResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.UploadURL != mock.url {
		t.Errorf("upload_url mismatch: got %q", got.UploadURL)
	}
	if got.ObjectKey == "" {
		t.Error("object_key should not be empty")
	}
}

func TestPresign_ValidationErrors(t *testing.T) {
	mock := &mockUploader{url: "http://rustfs/presigned"}
	srv := newTestServer(t, mock)
	defer srv.Close()

	cases := []struct {
		name string
		body string
	}{
		{"missing filename", `{"content_type":"image/jpeg","content_length":100}`},
		{"missing content_type", `{"filename":"f.jpg","content_length":100}`},
		{"zero content_length", `{"filename":"f.jpg","content_type":"image/jpeg","content_length":0}`},
		{"negative content_length", `{"filename":"f.jpg","content_type":"image/jpeg","content_length":-1}`},
		{"exceeds max size", `{"filename":"f.jpg","content_type":"image/jpeg","content_length":6000000000}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(srv.URL+"/upload/presign", "application/json", bytes.NewBufferString(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("want 422, got %d", resp.StatusCode)
			}
		})
	}
}

func TestPresign_BadJSON(t *testing.T) {
	srv := newTestServer(t, &mockUploader{})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/upload/presign", "application/json", strings.NewReader("not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestPresign_ObjectKeySanitisation(t *testing.T) {
	mock := &mockUploader{url: "http://rustfs/url", expiresAt: time.Now().Add(time.Minute)}
	srv := newTestServer(t, mock)
	defer srv.Close()

	// A path-traversal attempt in the filename should be neutralised.
	body := `{"filename":"../../etc/passwd","content_type":"text/plain","content_length":42}`
	resp, err := http.Post(srv.URL+"/upload/presign", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got handler.PresignResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(got.ObjectKey, "..") {
		t.Errorf("object_key contains path traversal: %q", got.ObjectKey)
	}
}

// --------------------------------------------------------------------------
// To run all tests in powershell: "go test -v ./metadata-server/..."
