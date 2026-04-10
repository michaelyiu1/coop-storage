package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strings"
	"time"
)

// ── Sentinel errors ───────────────────────────────────────────────────────────

var ErrNotFound = errors.New("object not found")
var ErrInternal = errors.New("internal storage error")

// ── Shared types ──────────────────────────────────────────────────────────────

// ObjectMeta holds the metadata record for a stored object.
type ObjectMeta struct {
	Key         string    `json:"key"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// ListFilter controls which objects are returned by ListObjects.
type ListFilter struct {
	Prefix string
}

// ── Interfaces ────────────────────────────────────────────────────────────────

// Uploader is the storage capability needed by UploadHandler.
type Uploader interface {
	PresignUpload(ctx context.Context, objectKey, contentType string, contentLength int64) (url string, expiresAt time.Time, err error)
}

// Deleter is the storage capability needed by DeleteHandler.
type Deleter interface {
	DeleteObject(ctx context.Context, key string) error
}

// Downloader is the storage capability needed by DownloadHandler.
type Downloader interface {
	PresignDownload(ctx context.Context, key string) (url string, expiresAt time.Time, err error)
}

// MetadataStore is the storage capability needed by MetadataHandler.
type MetadataStore interface {
	GetObject(ctx context.Context, key string) (*ObjectMeta, error)
	ListObjects(ctx context.Context, filter ListFilter) ([]ObjectMeta, error)
}

// ── JSON helpers ──────────────────────────────────────────────────────────────

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ── Shared validation ─────────────────────────────────────────────────────────

// validateObjectKey rejects empty keys and path-traversal attempts.
func validateObjectKey(key string) error {
	if key == "" {
		return errors.New("object key is required")
	}
	if strings.Contains(key, "..") {
		return errors.New("object key contains invalid path components")
	}
	if strings.Contains(path.Clean("/"+key), "..") {
		return errors.New("object key contains invalid path components")
	}
	return nil
}
