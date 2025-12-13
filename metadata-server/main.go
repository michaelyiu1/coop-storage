package main
import (
	"net/http"
	"fmt"
	"log"
	"encoding/json"
	"io"
)

// TODO: figure out cleaner way to share types across containers?
type MetadataPOST struct {
	ID string `json:"ID"`
	FileType string `json:"FileType"`
	FileName string `json:"FileName"`
}


// client -> server (TODO: unused)
type ReadFilter struct {
	Query string `json:"query"`
	FileType string `json:"FileType"`
}

func main() {
	InitDb()
	defer CloseDb()
	// client facing
	http.HandleFunc("/write_object", requestWriteObject) // maybe this one is just auth?
	http.HandleFunc("/get_meta", getMetaForUser)
	
	// called by osd
	http.HandleFunc("/write_meta", createMetaObject)
	http.HandleFunc("/update_meta", UpdateMetaObject)
	// dev only, reads full object
	http.HandleFunc("/read_meta", readMetaObject)
	log.Printf("Server starting on PORT %s\n", PORT)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil); err != nil {
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

func getMetaForUser(w http.ResponseWriter, r *http.Request) {
	//TODO: consume an API token to verify access
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// TODO: use ReadFilter
	page := 0 //r.URL.Query().Get("page")

	// TODO: get this from auth
	userTMP := r.URL.Query().Get("user")

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "Failed to read request body", http.StatusBadRequest)
	// 	return
	// }
	// defer r.Body.Close()
	// if err := json.Unmarshal(body, &filter); err != nil {
	// 	http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
	// 	return
	// }

	var rr ReadRequest
	rr.PageNum = page

	rr.Read(userTMP)
	jo, err := json.Marshal(rr)
	if err != nil {
		http.Error(w, "could not marshal object", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jo)
}

func UpdateMetaObject(w http.ResponseWriter, r *http.Request) {
	//TODO: consume an API token to verify access
	if r.Method != http.MethodGet {
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

	if erro := DBInst.UpdateObj(&currMeta); erro != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	
	// TODO: check if here we also need concurrency protection
	if currMeta.DeleteFlag {
		DBInst.Delete([]byte(currMeta.ID))
		UpdateUserIndex(currMeta.Owner, currMeta.ID, Remove)
	}

	currMeta.Write()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

//called by the OSD Server
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

	if err := metaObject.Write(); err != nil {
		http.Error(w, "Failed to write object", http.StatusInternalServerError)
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