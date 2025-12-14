package main
import (
	"fmt"
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"log"
	"iter"	
)

// TODO: don;t save ID in column, waste of space
type MetaObject struct {
	// KEY
	ID         string `json:"id"`
	//VALUES
	Owner      string `json:"owner"`
	FileType   string `json:"fileType"`
	FileName   string `json:"fileName"`
	DeleteFlag bool   `json:"deleteFlag"`
	Version    int 	  `json:"version"`
	// TODO: implement these for file integrity checks and multipart upload
	// offset int32
	// length int32
}

func (o *MetaObject) Read() (error) {
	if (o.ID == "") {
		return fmt.Errorf("MetaObject needs an id to be Read")
	}
	oKey := []byte(fmt.Sprintf("objid:%s", o.ID))
	log.Printf("MetaObject.Read: ID=%s, searching for key: %s\n", o.ID, string(oKey))
	// TODO: this seems very wrong, try to deserialize directly to self instead, given
	var currMeta MetaObject
	currMetaJSON, err := DBInst.Read(oKey)
	if err != nil {
		log.Printf("MetaObject.Read: Failed to read key %s: %v\n", string(oKey), err)
		return err
	}

	errj := json.Unmarshal([]byte(currMetaJSON), &currMeta) 

	if errj != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	o.Owner  = currMeta.Owner
	o.FileType = currMeta.FileType
	o.FileName = currMeta.FileName
	o.DeleteFlag = currMeta.DeleteFlag
	
	return nil
}

func (o *MetaObject) Create() (error) {
	

	//TODO: check that file name is unique for the user!


	// object index
	jsonStr, err := json.Marshal(o)
	if err != nil {
		log.Printf("MetaObject.Write: Failed to marshal object: %v\n", err)
		return err
	}
	oKey := fmt.Sprintf("objid:%s", o.ID)
	log.Printf("MetaObject.Write: ID=%s, writing key: %s\n", o.ID, oKey)
	if err := DBInst.Update([]byte(oKey), jsonStr); err != nil {
		log.Printf("MetaObject.Write: Failed to update key %s: %v\n", oKey, err)
		return err
	}

	if err:=UpdateUserIndex(o.Owner, o.ID, Add); err != nil {
		log.Printf("Error update user index, %v", err)
		return err
	}


	return nil
}

func StartGarbageCollector() (error) {
	DBInst.db.RunValueLogGC(0.5)
	return nil
}


type ObjPair struct {
    FileName   string
    ID string
}

func IDNameTupleIterator(user string) iter.Seq2[ObjPair, error] {
	return func(yield func(ObjPair, error) bool) {
		uKey := []byte(fmt.Sprintf("user:%s", user))
		
		currFiles :=  NewSet[string]()
		currFilesJSON, err := DBInst.Read(uKey)
		if err == badger.ErrKeyNotFound {
			currFilesJSON = []byte("{}")
		}
		
		if err != nil {
			log.Printf("OSDGuide.Read: Failed to read key %s: %v\n", user, err)
			yield(ObjPair{}, err)
			return
		}
		
		errj := json.Unmarshal([]byte(currFilesJSON), &currFiles) 
		
		if errj != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			yield(ObjPair{}, ErrOnWrite)
			return
		}
		
		if currFiles.Size() == 0 {
			yield(ObjPair{}, nil)
			return
		}
		
		// fish the metadata objects to get the filenames
		// TODO: use goroutines to speed up
		for key, _ := range(currFiles) {
			var currMeta MetaObject
			primaryKey := []byte(fmt.Sprintf("objid:%s", key))
			// TODO: again this should be currMeta.Read
			currMetaJSON, err := DBInst.Read(primaryKey)
			if err != nil {
				// log.Printf("OSDGuide.Read: Failed to read key %s: %v\n", string(key), err)
				yield(ObjPair{}, fmt.Errorf("OSDGuide.Read: Failed to read key %s: %v\n", string(primaryKey), err))
				return
			}
		
			errj := json.Unmarshal([]byte(currMetaJSON), &currMeta) 
		
			if errj != nil {
				yield(ObjPair{}, fmt.Errorf("OSDGuide.Read: Error unmarshalling JSON: %v", errj))
				return
			}
	
			if !yield(ObjPair{currMeta.FileName, string(key)}, nil) {
				return
			}
		}
	
	}
}

// UTILS
// user index stuff
type UpdateArrayMode int

const  (
	Add UpdateArrayMode = iota
	Remove
)

func UpdateUserIndex(user string, objId string, mode UpdateArrayMode) (error) {
	uKey := []byte(fmt.Sprintf("user:%s", user))
	currFiles :=  NewSet[string]()
	currFilesJSON, err := DBInst.Read(uKey)
	if err == badger.ErrKeyNotFound {
		currFilesJSON = []byte("{}")
	}

	errj := json.Unmarshal([]byte(currFilesJSON), &currFiles) 

	if errj != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return fmt.Errorf("UpdateUserIndex Error unmarshalling JSON")
	}

	switch mode {
	case Add:
		currFiles.Add(objId)
	case Remove:
		currFiles.Remove(objId)

	}

	currFilesJSON, err = json.Marshal(currFiles)
	DBInst.Update(uKey, currFilesJSON)
	
	return nil
}