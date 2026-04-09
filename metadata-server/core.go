package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
)

// TODO: don;t save ID in column, waste of space
type MetaObject struct {
	// KEY
	ID string `json:"id"`
	//VALUES
	Owner      string `json:"owner"`
	FileType   string `json:"fileType"`
	FileName   string `json:"fileName"`
	DeleteFlag bool   `json:"deleteFlag"`
	Version    int    `json:"version"`
	// TODO: implement these for file integrity checks and multipart upload
	// offset int32
	// length int32
}

func (o *MetaObject) Read() error {
	if o.ID == "" {
		return fmt.Errorf("MetaObject needs an id to be Read")
	}
	oKey := NewDBKey(Object, o.ID)
	// TODO: this seems very wrong, try to deserialize directly to self instead, given
	var currMeta MetaObject
	currMetaJSON, err := DBInst.Read(oKey)
	if err != nil {
		log.Printf("MetaObject.Read: Failed to read key %s: %v\n", string(oKey), err)
		return err
	}

	if err := json.Unmarshal([]byte(currMetaJSON), &currMeta); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	o.Owner = currMeta.Owner
	o.FileType = currMeta.FileType
	o.FileName = currMeta.FileName
	o.DeleteFlag = currMeta.DeleteFlag

	return nil
}

func (o *MetaObject) Create() error {
	// object index
	jsonStr, err := json.Marshal(o)
	if err != nil {
		log.Printf("MetaObject.Create: Failed to marshal object: %v\n", err)
		return err
	}
	oKey := NewDBKey(Object, o.ID)
	log.Printf("MetaObject.Create: ID=%s, writing key: %s\n", o.ID, oKey)

	if err := UpdateUserIndex(o.Owner, o.FileName, o.ID, "", Add); err != nil {
		log.Printf("MetaObject.Create Error update user index, %v", err)
		return err
	}

	// TODO: use newer UpdateObject instead
	if err := DBInst.Update([]byte(oKey), jsonStr); err != nil {
		log.Printf("MetaObject.Create: Failed to update key %s: %v\n", oKey, err)
		return err
	}

	return nil
}

func (o *MetaObject) Update() error {
	// TODO: check if here we also need concurrency protection
	if o.DeleteFlag {
		log.Println("Delete detected")
		if err := DBInst.Delete(NewDBKey(Object, o.ID)); err != nil {
			return err
		}
		if err := UpdateUserIndex(o.Owner, o.FileName, "", "", Remove); err != nil {
			return err
		}

		return nil
	}

	oKey := NewDBKey(Object, o.ID)
	var currMeta MetaObject
	currMetaJSON, err := DBInst.Read(oKey)
	if err != nil {
		log.Printf("MetaObject.Update: Failed to read key %s: %v\n", string(oKey), err)
		return err
	}

	if err := json.Unmarshal([]byte(currMetaJSON), &currMeta); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	//NOTE: done first to check for uniqueness
	if err := UpdateUserIndex(o.Owner, o.FileName, o.ID, currMeta.FileName, Modify); err != nil {
		return err
	}

	if err := DBInst.UpdateObj(o); err != nil {
		log.Printf("failed to update %v", err)
		return err
	}

	return nil
}

func StartGarbageCollector() error {
	DBInst.db.RunValueLogGC(0.5)
	return nil
}

////////////
// UTILS  //
////////////

// user index stuff
type UpdateArrayMode int

const (
	Add UpdateArrayMode = iota
	Modify
	Remove
)

// always call this first to ensure fName uniqueness
// TODO: concurrency protection?
func UpdateUserIndex(user string, fName string, objId string, oldFname string, mode UpdateArrayMode) error {
	uKey := NewDBKey(User, user)
	objectMap := make(map[string]string)
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

	switch mode {
	case Add:
		if _, ok := objectMap[fName]; ok {
			return fmt.Errorf("UpdateUserIndex: file name already exists for user")
		}
		objectMap[fName] = objId
	case Modify:
		if _, ok := objectMap[fName]; ok {
			return fmt.Errorf("UpdateUserIndex: file name already exists for user")
		}
		delete(objectMap, oldFname)
		objectMap[fName] = objId
	case Remove:
		delete(objectMap, fName)
	}

	objectMapJSON, err = json.Marshal(objectMap)
	DBInst.Update(uKey, objectMapJSON)

	return nil
}
