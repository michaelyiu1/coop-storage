package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	handler "github.com/bfbarry/coop-storage/metadata-server/controllers"
)

// mockDownloader lets tests control what the storage layer returns for presign-get.
type mockDownloader struct {
	url       string
	expiresAt time.Time
	err       error
}

func (m *mockDownloader) PresignDownload(_ context.Context, _ string) (string, time.Time, error) {
	return m.url, m.expiresAt, m.err
}

func newDownloadTestServer(t *testing.T, dl handler.Downloader) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handler.NewDownloadHandler(dl).Register(mux)
	return httptest.NewServer(mux)
}

// ---------------------------------------------------------------------------
// Happy path
// ---------------------------------------------------------------------------

func TestPresignDownload_HappyPath(t *testing.T) {
	expiry := time.Now().Add(15 * time.Minute)
	mock := &mockDownloader{url: "http://rustfs/presigned-get-url", expiresAt: expiry}
	srv := newDownloadTestServer(t, mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/my-object-key")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var got handler.PresignDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.DownloadURL != mock.url {
		t.Errorf("download_url mismatch: got %q", got.DownloadURL)
	}
	if got.ExpiresAt.IsZero() {
		t.Error("expires_at should not be zero")
	}
}

// ---------------------------------------------------------------------------
// Object not found
// ---------------------------------------------------------------------------

func TestPresignDownload_ObjectNotFound(t *testing.T) {
	mock := &mockDownloader{err: handler.ErrNotFound}
	srv := newDownloadTestServer(t, mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/ghost-key")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Storage error
// ---------------------------------------------------------------------------

func TestPresignDownload_StorageError(t *testing.T) {
	mock := &mockDownloader{err: handler.ErrInternal}
	srv := newDownloadTestServer(t, mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/some-key")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Missing key in path
// ---------------------------------------------------------------------------

func TestPresignDownload_MissingKey(t *testing.T) {
	srv := newDownloadTestServer(t, &mockDownloader{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Path traversal
// ---------------------------------------------------------------------------

func TestPresignDownload_PathTraversal(t *testing.T) {
	srv := newDownloadTestServer(t, &mockDownloader{url: "http://rustfs/url", expiresAt: time.Now().Add(time.Minute)})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/..%2F..%2Fetc%2Fpasswd")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for path traversal, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Expiry window
// ---------------------------------------------------------------------------

func TestPresignDownload_ExpiryIsInFuture(t *testing.T) {
	expiry := time.Now().Add(15 * time.Minute)
	mock := &mockDownloader{url: "http://rustfs/url", expiresAt: expiry}
	srv := newDownloadTestServer(t, mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/my-key")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got handler.PresignDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.ExpiresAt.After(time.Now()) {
		t.Errorf("expires_at should be in the future, got %v", got.ExpiresAt)
	}
}

// ---------------------------------------------------------------------------
// Wrong HTTP methods
// ---------------------------------------------------------------------------

func TestPresignDownload_WrongMethod(t *testing.T) {
	srv := newDownloadTestServer(t, &mockDownloader{})
	defer srv.Close()

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, srv.URL+"/download/presign/key", nil)
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

// ---------------------------------------------------------------------------
// Content-Type
// ---------------------------------------------------------------------------

func TestPresignDownload_ResponseContentTypeIsJSON(t *testing.T) {
	mock := &mockDownloader{url: "http://rustfs/url", expiresAt: time.Now().Add(time.Minute)}
	srv := newDownloadTestServer(t, mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download/presign/key")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("want application/json Content-Type, got %q", ct)
	}
}
