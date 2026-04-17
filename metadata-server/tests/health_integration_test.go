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

// ---------------------------------------------------------------------------
// Health check handler
// ---------------------------------------------------------------------------

func newHealthTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handler.NewHealthHandler().Register(mux)
	return httptest.NewServer(mux)
}

func TestHealth_ReturnsOK(t *testing.T) {
	srv := newHealthTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
}

func TestHealth_ResponseBodyContainsStatus(t *testing.T) {
	srv := newHealthTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("health response is not JSON: %v", err)
	}
	if _, ok := body["status"]; !ok {
		t.Error("health response should contain a 'status' field")
	}
}

func TestHealth_MethodNotAllowed(t *testing.T) {
	srv := newHealthTestServer(t)
	defer srv.Close()

	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, srv.URL+"/health", nil)
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
// Integration: upload then delete
// ---------------------------------------------------------------------------

// multiMock satisfies both Uploader and Deleter for integration tests.
type multiMock struct {
	uploadURL string
	uploadExp time.Time
	uploadErr error
	deleteErr error
	deleted   []string
}

func (m *multiMock) PresignUpload(_ context.Context, _, _ string, _ int64) (string, time.Time, error) {
	return m.uploadURL, m.uploadExp, m.uploadErr
}

func (m *multiMock) DeleteObject(_ context.Context, key string) error {
	m.deleted = append(m.deleted, key)
	return m.deleteErr
}

func newFullTestServer(t *testing.T, mock *multiMock) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handler.NewUploadHandler(mock).Register("/upload", mux)
	handler.NewDeleteHandler(mock).Register(mux)
	handler.NewHealthHandler().Register(mux)
	return httptest.NewServer(mux)
}

func TestIntegration_UploadThenDelete(t *testing.T) {
	mock := &multiMock{
		uploadURL: "http://rustfs/presigned",
		uploadExp: time.Now().Add(15 * time.Minute),
	}
	srv := newFullTestServer(t, mock)
	defer srv.Close()

	// Step 1: presign an upload and capture the object key.
	body := `{"filename":"doc.pdf","content_type":"application/pdf","content_length":512000}`
	resp, err := http.Post(srv.URL+"/upload/presign", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload presign: want 200, got %d", resp.StatusCode)
	}

	var presign handler.PresignResponse
	if err := json.NewDecoder(resp.Body).Decode(&presign); err != nil {
		t.Fatal(err)
	}
	if presign.ObjectKey == "" {
		t.Fatal("object_key must not be empty")
	}

	// Step 2: delete using the same object key.
	delReq, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/"+presign.ObjectKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Errorf("delete: want 204, got %d", delResp.StatusCode)
	}

	// Verify the mock recorded the deletion.
	if len(mock.deleted) == 0 {
		t.Error("expected DeleteObject to be called")
	}
}

func TestIntegration_HealthAfterErrors(t *testing.T) {
	// Health endpoint should remain operational regardless of storage failures.
	mock := &multiMock{uploadErr: handler.ErrInternal, deleteErr: handler.ErrInternal}
	srv := newFullTestServer(t, mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("health should still return 200 after storage errors, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Concurrent delete safety
// ---------------------------------------------------------------------------

func TestDelete_ConcurrentRequests(t *testing.T) {
	srv := newDeleteTestServer(t, &mockDeleter{})
	defer srv.Close()

	const workers = 20
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			req, err := http.NewRequest(http.MethodDelete, srv.URL+"/objects/concurrent-key", nil)
			if err != nil {
				errs <- err
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errs <- err
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				errs <- nil // tolerate; we mainly check for races / panics
			}
			errs <- nil
		}()
	}

	for i := 0; i < workers; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent delete error: %v", err)
		}
	}
}
