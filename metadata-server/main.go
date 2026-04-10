package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bfbarry/coop-storage/metadata-server/config"
	"github.com/bfbarry/coop-storage/metadata-server/controllers"
	"github.com/bfbarry/coop-storage/metadata-server/storage"
)

// TODO: figure out cleaner way to share types across containers?
type MetadataPOST struct {
	ID       string `json:"ID"`
	FileType string `json:"FileType"`
	FileName string `json:"FileName"`
}

// client -> server (TODO: unused)
type ReadFilter struct {
	Query    string `json:"query"`
	FileType string `json:"FileType"`
}

func main() {
	if config.ISDEV {
		log.SetFlags(0)
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

	uploader.Register("/upload/presign", mux)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, World!")
		if err != nil {
			panic("ahhh")
		}
	})
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

	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), mux); err != nil {
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
	metaObject.Owner = "placeholder" // TODO: get this from auth
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

func runGc(w http.ResponseWriter, r *http.Request) {
	StartGarbageCollector()
	log.Printf("garbage collection ran")
	w.Write([]byte("ok"))
}
