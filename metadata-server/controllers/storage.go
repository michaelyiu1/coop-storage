package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HTTP I/O

// PresignRequest is the JSON body the client sends.
type PresignRequest struct {
	// Filename is used only to derive the object key prefix; the actual
	// storage key is always UUID-based so filenames can never collide.
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	// ContentLength must be the exact byte size of the file to upload.
	// RustFS will reject the PUT if the value doesn't match.
	ContentLength int64 `json:"content_length"`
}

// PresignResponse is returned to the client on success.
type PresignResponse struct {
	// UploadURL is the pre-signed PUT URL. The client must:
	//   PUT <file bytes>  →  UploadURL
	//   Headers required:  Content-Type: <original content_type>
	//                      Content-Length: <original content_length>
	UploadURL string `json:"upload_url"`
	// ObjectKey is the storage path. Persist this in your DB so you can
	// generate download URLs or delete the object later.
	ObjectKey string `json:"object_key"`
	// ExpiresAt tells the client how long the URL is valid.
	ExpiresAt time.Time `json:"expires_at"`
}

// Uploader is the storage capability this handler needs.
// Keeping it as an interface makes the handler trivially testable.
type Uploader interface {
	PresignUpload(ctx context.Context, objectKey, contentType string, contentLength int64) (url string, expiresAt time.Time, err error)
}

type UploadHandler struct {
	store Uploader
}

func NewUploadHandler(store Uploader) *UploadHandler {
	return &UploadHandler{store: store}
}

func (h *UploadHandler) handlePresign(w http.ResponseWriter, r *http.Request) {
	var req PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validatePresignRequest(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// TODO(auth): replace "anonymous" with the authenticated user's ID once
	// auth middleware is wired up. The user ID scopes the object key so each
	// user's files are isolated within the shared bucket.
	userID := "anonymous"
	objectKey := buildObjectKey(userID, req.Filename)

	uploadURL, expiresAt, err := h.store.PresignUpload(r.Context(), objectKey, req.ContentType, req.ContentLength)
	if err != nil {
		// Don't leak internal storage errors to the client.
		http.Error(w, "failed to generate upload URL", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, PresignResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresAt: expiresAt,
	})
}

// buildObjectKey constructs a deterministic, collision-free storage path.
//
// Format:  {userID}/{fileID}/{sanitisedFilename}
//
// Keeping the original filename as a suffix preserves the content-type hint
// for tooling that inspects the key, while the UUID guarantees uniqueness.
func buildObjectKey(userID, filename string) string {
	fileID := uuid.New().String()
	safe := sanitiseFilename(filename)
	return fmt.Sprintf("%s/%s/%s", userID, fileID, safe)
}

// sanitiseFilename strips path separators so a crafted filename can't escape
// the user's storage namespace (e.g. "../../other-user/secret").
func sanitiseFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" {
		return "file"
	}
	return name
}

func validatePresignRequest(req PresignRequest) error {
	var errs []string

	if strings.TrimSpace(req.Filename) == "" {
		errs = append(errs, "filename is required")
	}
	if strings.TrimSpace(req.ContentType) == "" {
		errs = append(errs, "content_type is required")
	}
	if req.ContentLength <= 0 {
		errs = append(errs, "content_length must be a positive integer (bytes)")
	}

	const maxBytes = 5 * 1024 * 1024 * 1024 // 5 GB — tune to your plan limits
	if req.ContentLength > maxBytes {
		errs = append(errs, fmt.Sprintf("content_length exceeds maximum allowed size of %d bytes", maxBytes))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

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
