package main
import (
	"fmt"
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"log"

)
// In this module: logic to interface between the client and the metadata server

// server -> client payload
type ReadRequest struct {
	Token string `json:"Token"`
	PageNum int `json:"PageNum"` //TODO: implement paging
	IDMap map[string]string `json:"IDMap"`
}

func (self *ReadRequest) Read(user string) (error) {
	uKey := []byte(fmt.Sprintf("user:%s", user))
	
	var currFiles []string
	currFilesJSON, err := DBInst.Read(uKey)
	if err == badger.ErrKeyNotFound {
		currFilesJSON = []byte("[]")
	}

	if err != nil {
		log.Printf("ReadRequest.Read: Failed to read key %s: %v\n", user, err)
		return err
	}

	errj := json.Unmarshal([]byte(currFilesJSON), &currFiles) 

	if errj != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	self.IDMap = make(map[string]string)
	if len(currFiles) == 0 {
		return nil
	}

	// fish the metadata objects to get the filenames
	// TODO: use goroutines to speed up
	for _, val := range(currFiles) {
		var currMeta MetaObject
		key := []byte(fmt.Sprintf("objid:%s", val))
		// TODO: again this should be currMeta.Read
		currMetaJSON, err := DBInst.Read(key)
		if err != nil {
			log.Printf("ReadRequest.Read: Failed to read key %s: %v\n", string(key), err)
			return err
		}
	
		errj := json.Unmarshal([]byte(currMetaJSON), &currMeta) 
	
		if errj != nil {
			fmt.Println("ReadRequest.Read: Error unmarshalling JSON:", err)
			return ErrOnWrite
		}

		self.IDMap[currMeta.FileName] = val
	}
	return nil
}
