package controllers

import (
	"errors"
	"net/http"
	"strings"
)

type DeleteHandler struct {
	store Deleter
}

func NewDeleteHandler(store Deleter) *DeleteHandler {
	return &DeleteHandler{store: store}
}

func (h *DeleteHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/objects/", h.handle)
}

func (h *DeleteHandler) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/objects/")
	if err := validateObjectKey(key); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.DeleteObject(r.Context(), key); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete object")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
