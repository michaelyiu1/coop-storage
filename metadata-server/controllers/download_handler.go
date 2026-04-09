package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// PresignDownloadResponse is returned to the client by the download presign endpoint.
type PresignDownloadResponse struct {
	DownloadURL string    `json:"download_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type DownloadHandler struct {
	store Downloader
}

func NewDownloadHandler(store Downloader) *DownloadHandler {
	return &DownloadHandler{store: store}
}

func (h *DownloadHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/download/presign/", h.handle)
}

func (h *DownloadHandler) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/download/presign/")
	if err := validateObjectKey(key); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	downloadURL, expiresAt, err := h.store.PresignDownload(r.Context(), key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to generate download URL")
		return
	}

	writeJSON(w, http.StatusOK, PresignDownloadResponse{
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt,
	})
}
