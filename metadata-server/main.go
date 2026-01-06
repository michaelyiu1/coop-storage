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
	if ISDEV {
		log.SetFlags(0)
	}
	InitDb() //sets up the database connection
	defer CloseDb() //ensures the database is cleanly closed when the program exits, runs even if server crashes later
	// client facing
	http.HandleFunc("/write_object", requestWriteObject) // maybe this one is just auth?
	http.HandleFunc("/prepare_osd_request", prepareOSDRequest)
	
	// called by osd
	http.HandleFunc("/write_meta", createMetaObject)
	http.HandleFunc("/update_meta", UpdateMetaObject)
	// dev only
	http.HandleFunc("/read_meta", readMetaObject)
	http.HandleFunc("/run_gc", runGc)
	log.Printf("Server starting on PORT %s\n", PORT)
	
	// if err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil); err != nil {
	// 	log.Fatal("Server failed to start:", err)
	// }

	    // 1️⃣ Create mux
    mux := http.NewServeMux()

    // 2️⃣ Register handlers
    mux.HandleFunc("/read_meta", readMetaObject)
    mux.HandleFunc("/write_meta", createMetaObject)
    mux.HandleFunc("/update_meta", UpdateMetaObject)
	mux.HandleFunc("/list_meta_ids", listMetaIDs)
    // add other handlers here

    // 3️⃣ Wrap mux with CORS middleware when starting the server
    log.Printf("Server starting on PORT %s\n", PORT)
    if err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), enableCORS(mux)); err != nil {
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

// Helps client retrieve or update objects on OSD
func prepareOSDRequest(w http.ResponseWriter, r *http.Request) {
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
	var og OSDGuide
	og.PageNum = page
	//TODO: if user not found, return 404
	if err := og.Read(userTMP); err == KeyNotFound {
		http.Error(w, fmt.Sprintf("key not found , %v", userTMP), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("could not read : %v", err), http.StatusInternalServerError)

	}

	jo, err := json.Marshal(og)
	if err != nil {
		http.Error(w, "could not marshal object", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jo)
}

//TODO: shouldn't be able to edit things like owner or filetype,
//so we shall create Base objects w/ composition to make type definition easier
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

func createMetaObject(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        log.Printf("Rejected non-POST request: %s", r.Method)
        return
    }

    log.Println("createMetaObject invoked")

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        log.Printf("Error reading request body: %v", err)
        return
    }
    defer r.Body.Close()

    log.Printf("Request body: %s", string(body))

    var metaPost MetadataPOST
    if err := json.Unmarshal(body, &metaPost); err != nil {
        http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
        log.Printf("JSON unmarshal error: %v", err)
        return
    }

    log.Printf("Parsed MetadataPOST: %+v", metaPost)

    if DBInst == nil || DBInst.db == nil {
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        log.Printf("DBInst is nil")
        return
    }

    var metaObject MetaObject
    metaObject.ID = metaPost.ID
    metaObject.FileType = metaPost.FileType
    metaObject.FileName = metaPost.FileName
    metaObject.Owner = "placeholder" // TODO: get from auth
    metaObject.DeleteFlag = false

    log.Printf("Creating MetaObject: %+v", metaObject)

    if err := metaObject.Create(); err != nil {
        http.Error(w, fmt.Sprintf("Failed to create object %v", err), http.StatusInternalServerError)
        log.Printf("metaObject.Create failed: %v", err)
        return
    }

    log.Printf("MetaObject created successfully: ID=%s", metaObject.ID)

    // Return JSON response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "id":     metaObject.ID,
    })
}

// //called by the OSD Server
// func createMetaObject(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
	
// 	log.Printf("createMetaObject invoked")
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Failed to read request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()
	
// 	var metaPost MetadataPOST
// 	if err := json.Unmarshal(body, &metaPost); err != nil {
// 		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
// 		return
// 	}
	
// 	var metaObject MetaObject
// 	metaObject.ID = metaPost.ID
// 	metaObject.FileType = metaPost.FileType
// 	metaObject.FileName = metaPost.FileName
// 	metaObject.Owner = "placeholder" // TODO: get this from auth
// 	metaObject.DeleteFlag = false
	

// 	if err := metaObject.Create(); err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to create object %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	fmt.Fprint(w, "success")

// 	//return JSON and set headers
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)

// 	json.NewEncoder(w).Encode(map[string]string{
// 		"status": "success",
// 		"id": metaObject.ID,
// 	})

// }

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

//List all Meta IDs
func listMetaIDs(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ids, err := getAllMetaIDs() // function that returns []string of all object IDs
    if err != nil {
        http.Error(w, "Failed to fetch IDs", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ids)
}


func runGc(w http.ResponseWriter, r *http.Request) {
	StartGarbageCollector()
	log.Printf("garbage collection ran")
	w.Write([]byte("ok"))
}

//CORS to enable vue dev server to access data
func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Vue dev server
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}