package controllers

import (
	"errors"
	"net/http"
	"strings"
)

type MetadataHandler struct {
	store MetadataStore
}

func NewMetadataHandler(store MetadataStore) *MetadataHandler {
	return &MetadataHandler{store: store}
}

func (h *MetadataHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/objects", h.handleList)
	mux.HandleFunc("/objects/", h.handleGet)
}

func (h *MetadataHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/objects/")
	if err := validateObjectKey(key); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	obj, err := h.store.GetObject(r.Context(), key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve object metadata")
		return
	}

	writeJSON(w, http.StatusOK, obj)
}

func (h *MetadataHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	filter := ListFilter{Prefix: r.URL.Query().Get("prefix")}

	objects, err := h.store.ListObjects(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list objects")
		return
	}

	if objects == nil {
		objects = []ObjectMeta{}
	}

	writeJSON(w, http.StatusOK, objects)
}
