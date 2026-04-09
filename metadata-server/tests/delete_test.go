package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/bfbarry/coop-storage/metadata-server/controllers"
)

// mockDeleter lets tests control what the storage layer returns for delete ops.
type mockDeleter struct {
	err error
}

func (m *mockDeleter) DeleteObject(_ context.Context, _ string) error {
	return m.err
}

func newDeleteTestServer(t *testing.T, deleter handler.Deleter) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handler.NewDeleteHandler(deleter).Register(mux)
	return httptest.NewServer(mux)
}

// ---------------------------------------------------------------------------
// Happy path
// ---------------------------------------------------------------------------

func TestDelete_HappyPath(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{})
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/abc123", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("want 204, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Missing / malformed object key
// ---------------------------------------------------------------------------

func TestDelete_MissingObjectKey(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{})
	defer srv.Close()

	// Route without a key segment should be rejected.
	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestDelete_PathTraversalInKey(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{})
	defer srv.Close()

	// URL-encoded path traversal attempt.
	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/..%2F..%2Fetc%2Fpasswd", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Handler must reject traversal attempts, not forward them to storage.
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400 for path traversal, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Storage-layer errors
// ---------------------------------------------------------------------------

func TestDelete_StorageNotFound(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{err: handler.ErrNotFound})
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/missing-key", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestDelete_StorageInternalError(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{err: handler.ErrInternal})
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/some-key", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Wrong HTTP methods
// ---------------------------------------------------------------------------

func TestDelete_WrongMethod(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{})
	defer srv.Close()

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, srv.URL+"/objects/abc123", nil)
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
// Response body
// ---------------------------------------------------------------------------

func TestDelete_ResponseBodyIsEmptyOnSuccess(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{})
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/abc123", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var buf [1]byte
	n, _ := resp.Body.Read(buf[:])
	if n != 0 {
		t.Errorf("expected empty body on 204, got %d byte(s)", n)
	}
}

func TestDelete_ErrorResponseIsJSON(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{err: handler.ErrNotFound})
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/missing", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("want Content-Type application/json, got %q", ct)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("error body is not valid JSON: %v", err)
	}
	if _, ok := payload["error"]; !ok {
		t.Error("JSON error body should contain an 'error' field")
	}
}
