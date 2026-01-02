package main

//import GO packages
import (
	"log" //library package for simple logging. Used to print messages with automatic timestamps to the console or file, which helps with debugging and tracking app behavior.
	"encoding/json" //Used for JSON operations. It allows you to convert Go data (like structs) into JSON format (called "marshalling") and turn JSON strings back into Go data ("unmarshalling")
	"fmt" //Short for "format." It is the most common package for formatted I/O. You use it to print text to the terminal (e.g., fmt.Println), format strings without printing them (fmt.Sprintf), or read user input (fmt.Scan).
	"github.com/dgraph-io/badger/v4" // library for BadgerDB, which is a fast, embeddable key-value database. It allows your program to store data persistently on a disk without needing a separate database server like MySQL or PostgreSQL
)
// In this module: logic to interface between the client and the metadata server

// server -> client payload
// custom datatype (aka class in python), token, pagenum, and IDmap are exported from the server
type OSDGuide struct {
	Token string `json:"Token"`
	PageNum int `json:"PageNum"` //TODO: implement paging
	IDMap map[string]string `json:"IDMap"`
}

func (self *OSDGuide) Read(user string) (error) {
	objectMap :=  make(map[string]string) // empty dictionary equivalent
	uKey := NewDBKey(User, user)
	objectMapJSON, err := DBInst.Read(uKey)
	if err == badger.ErrKeyNotFound {
		return KeyNotFound
	} else if err != nil {
		return err
	}

	// unpackage the json, if there is an error, print it
	if err := json.Unmarshal([]byte(objectMapJSON), &objectMap); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return fmt.Errorf("UpdateUserIndex Error unmarshalling JSON")
	}
	
	self.IDMap = objectMap
	return nil
}
