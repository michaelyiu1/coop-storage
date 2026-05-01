package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"os"

	"github.com/bfbarry/coop-storage/metadata-server/config"
	"github.com/bfbarry/coop-storage/metadata-server/controllers"
	"github.com/bfbarry/coop-storage/metadata-server/storage"
)

// TODO: figure out cleaner way to share types across containers?
type MetadataPOST struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	FileType string `json:"fileType"`
	FileName string `json:"fileName"`
}

// client -> server (TODO: unused)
type ReadFilter struct {
	Query    string `json:"query"`
	FileType string `json:"FileType"`
}

// CORS middleware to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if config.ISDEV {
		log.SetFlags(0)
		log.SetOutput(os.Stdout)
	}
	InitDb()
	defer CloseDb()
	// TODO: add more config to http server e.g,
	// 		Addr:         ":" + config.Server.Port,
	// Handler:      mux,
	// ReadTimeout:  10 * time.Second,
	// WriteTimeout: 10 * time.Second,
	// IdleTimeout:  60 * time.Second

	mux := http.NewServeMux()

	rustFsClient := storage.NewClient(config.GLOBAL_CONFIG.RustFS)

	uploader := controllers.NewUploadHandler(rustFsClient)
	downloader := controllers.NewDownloadHandler(rustFsClient)

	// uploader.Register("/upload/presign", mux)
	// downloader.Register("/download/presign", mux)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("/upload/presign", uploader.HandlePresign)
	mux.HandleFunc("/download/presign/", downloader.HandlePresign)
	mux.HandleFunc("/write_meta", createMetaObject)
	mux.HandleFunc("/read_meta", readMetaObject)
	mux.HandleFunc("/read_all_meta", readAllMetaObjects)

	// client facing
	// http.HandleFunc("/write_object", requestWriteObject) // maybe this one is just auth?
	// http.HandleFunc("/prepare_osd_request", uploader.)

	// // called by osd
	// http.HandleFunc("/write_meta", createMetaObject)
	// http.HandleFunc("/update_meta", UpdateMetaObject)
	// // dev only
	// http.HandleFunc("/read_meta", readMetaObject)
	// http.HandleFunc("/run_gc", runGc)
	log.Printf("Server starting on PORT %s\n", config.PORT)

	// Wrap mux with CORS middleware
	handler := corsMiddleware(mux)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// Called by client
func requestWriteObject(w http.ResponseWriter, r *http.Request) {
	//TODO: consume an API token to verify access
	// TODO: figure out if what other useful data this controller can return
	//  is really necessary
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("tokenplaceholder"))

}

// TODO: shouldn't be able to edit things like owner or filetype,
// so we shall create Base objects w/ composition to make type definition easier
func UpdateMetaObject(w http.ResponseWriter, r *http.Request) {
	//TODO: consume an API token to verify access
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var currMeta MetaObject
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &currMeta); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	err = currMeta.Update()
	if err != nil {
		http.Error(w, fmt.Sprintf("Method not allowed, %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// called by the OSD Server
func createMetaObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("createMetaObject invoked")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var metaPost MetadataPOST
	if err := json.Unmarshal(body, &metaPost); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	var metaObject MetaObject
	metaObject.ID = metaPost.ID
	metaObject.FileType = metaPost.FileType
	metaObject.FileName = metaPost.FileName
	metaObject.Owner = metaPost.Owner // TODO: get this from auth
	metaObject.DeleteFlag = false

	if err := metaObject.Create(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create object %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "success")
}

// For Dev Purposes
func readMetaObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	var metaObject MetaObject
	metaObject.ID = id
	if err := metaObject.Read(); err != nil {
		http.Error(w, fmt.Sprintf("Key objid:%s not found", id), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	o, err := json.Marshal(metaObject)
	if err != nil {
		http.Error(w, "could not marshal object", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(o)
}

// Read all metadata objects for a particular user
func readAllMetaObjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from query parameter
	// TODO: get this from auth token instead
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "user parameter is required", http.StatusBadRequest)
		return
	}

	// Read the user index to get all object IDs
	uKey := NewDBKey(User, user)
	objectMapJSON, err := DBInst.Read(uKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("User %s not found or has no objects", user), http.StatusNotFound)
		return
	}

	// Parse the user index (map of filename -> object ID)
	objectMap := make(map[string]string)
	if err := json.Unmarshal(objectMapJSON, &objectMap); err != nil {
		http.Error(w, "Failed to parse user index", http.StatusInternalServerError)
		return
	}

	// Retrieve all metadata objects for this user
	metaObjects := make([]MetaObject, 0, len(objectMap))
	for _, objID := range objectMap {
		var metaObject MetaObject
		metaObject.ID = objID
		if err := metaObject.Read(); err != nil {
			log.Printf("Warning: Failed to read object %s for user %s: %v", objID, user, err)
			continue // Skip objects that can't be read
		}
		metaObjects = append(metaObjects, metaObject)
	}

	// Return the array of metadata objects
	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(metaObjects)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func runGc(w http.ResponseWriter, r *http.Request) {
	StartGarbageCollector()
	log.Printf("garbage collection ran")
	w.Write([]byte("ok"))
}
