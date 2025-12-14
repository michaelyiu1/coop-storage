package main
import (
	"log"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
)
// In this module: logic to interface between the client and the metadata server

// server -> client payload
type OSDGuide struct {
	Token string `json:"Token"`
	PageNum int `json:"PageNum"` //TODO: implement paging
	IDMap map[string]string `json:"IDMap"`
}

func (self *OSDGuide) Read(user string) (error) {
	objectMap :=  make(map[string]string)
	uKey := NewDBKey(User, user)
	objectMapJSON, err := DBInst.Read(uKey)
	if err == badger.ErrKeyNotFound {
		objectMapJSON = []byte("{}")
	} else {
		return err
	}
		
	if err := json.Unmarshal([]byte(objectMapJSON), &objectMap); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return fmt.Errorf("UpdateUserIndex Error unmarshalling JSON")
	}
	
	self.IDMap = objectMap
	return nil
}
