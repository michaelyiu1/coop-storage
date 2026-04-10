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

// mockMetadataStore lets tests control list/get responses.
type mockMetadataStore struct {
	object *handler.ObjectMeta
	list   []handler.ObjectMeta
	err    error
}

func (m *mockMetadataStore) GetObject(_ context.Context, _ string) (*handler.ObjectMeta, error) {
	return m.object, m.err
}

func (m *mockMetadataStore) ListObjects(_ context.Context, _ handler.ListFilter) ([]handler.ObjectMeta, error) {
	return m.list, m.err
}

func newMetadataTestServer(t *testing.T, store handler.MetadataStore) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handler.NewMetadataHandler(store).Register(mux)
	return httptest.NewServer(mux)
}

// ---------------------------------------------------------------------------
// GET /objects/{key}
// ---------------------------------------------------------------------------

func TestGetObject_HappyPath(t *testing.T) {
	obj := &handler.ObjectMeta{
		Key:         "uploads/photo.jpg",
		ContentType: "image/jpeg",
		Size:        1048576,
		UploadedAt:  time.Now().UTC(),
	}
	srv := newMetadataTestServer(t, &mockMetadataStore{object: obj})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects/uploads%2Fphoto.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var got handler.ObjectMeta
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Key != obj.Key {
		t.Errorf("key mismatch: got %q, want %q", got.Key, obj.Key)
	}
	if got.Size != obj.Size {
		t.Errorf("size mismatch: got %d, want %d", got.Size, obj.Size)
	}
}

func TestGetObject_NotFound(t *testing.T) {
	srv := newMetadataTestServer(t, &mockMetadataStore{err: handler.ErrNotFound})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestGetObject_StorageError(t *testing.T) {
	srv := newMetadataTestServer(t, &mockMetadataStore{err: handler.ErrInternal})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects/some-key")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", resp.StatusCode)
	}
}

func TestGetObject_MissingKey(t *testing.T) {
	srv := newMetadataTestServer(t, &mockMetadataStore{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// GET /objects  (list)
// ---------------------------------------------------------------------------

func TestListObjects_HappyPath(t *testing.T) {
	objects := []handler.ObjectMeta{
		{Key: "a/file1.jpg", ContentType: "image/jpeg", Size: 100},
		{Key: "a/file2.png", ContentType: "image/png", Size: 200},
	}
	srv := newMetadataTestServer(t, &mockMetadataStore{list: objects})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	var got []handler.ObjectMeta
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(objects) {
		t.Errorf("want %d objects, got %d", len(objects), len(got))
	}
}

func TestListObjects_EmptyResult(t *testing.T) {
	srv := newMetadataTestServer(t, &mockMetadataStore{list: []handler.ObjectMeta{}})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 on empty list, got %d", resp.StatusCode)
	}

	var got []handler.ObjectMeta
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d item(s)", len(got))
	}
}

func TestListObjects_WithPrefixFilter(t *testing.T) {
	// The mock doesn't filter; we only verify the request is forwarded correctly
	// and a 200 is returned.
	srv := newMetadataTestServer(t, &mockMetadataStore{list: []handler.ObjectMeta{}})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects?prefix=uploads/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

func TestListObjects_StorageError(t *testing.T) {
	srv := newMetadataTestServer(t, &mockMetadataStore{err: handler.ErrInternal})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Content-Type header
// ---------------------------------------------------------------------------

func TestGetObject_ResponseContentTypeIsJSON(t *testing.T) {
	obj := &handler.ObjectMeta{Key: "k", ContentType: "image/jpeg", Size: 1}
	srv := newMetadataTestServer(t, &mockMetadataStore{object: obj})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/objects/k")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("want application/json Content-Type, got %q", ct)
	}
}
