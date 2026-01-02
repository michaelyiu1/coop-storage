package main

import (
	"log"
	"github.com/dgraph-io/badger/v4"
	"fmt"
	"encoding/json"
)

//DB wraps the Badger database
type DB struct{
	db *badger.DB
}

//DBInst is a global singleton
//You open the DB once
//Everything else uses DBInst
var DBInst *DB


//Creates Badger config (DB_PATH is where data lives on disk)
func InitDb() {

	opts := badger.DefaultOptions(DB_PATH)
	// opts.Logger = nil

	db, err := badger.Open(opts) //Opens the database
	DBInst = &DB {db} //Stores it in DBInst

	//Crashes the program if it fails
	if err != nil {
		log.Fatalf("Error opening BadgerDB: %v", err)
	}

}

func CloseDb() {
	if err := DBInst.db.Close(); err != nil {
		log.Printf("Error closing BadgerDB: %v", err)
	}
}

//Starts a write transaction, Writes key → val, Commits the transaction
Calls db.Sync() to flush to disk
func (self *DB) Update(key DBKey, val []byte) (error) {
	err := self.db.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(key, val))
		if err != nil {
			log.Printf("UPDATE: Error setting entry for key %s: %v", string(key), err)
			return err
		}
		log.Printf("UPDATE: Successfully set key: %s", string(key))
		return nil
	})

	if err != nil {
		log.Printf("UPDATE: Transaction failed for key %s: %v", string(key), err)
		return err
	}
	
	if err := self.db.Sync(); err != nil {
		log.Printf("UPDATE: Sync failed for key %s: %v", string(key), err)
		return err
	}
	
	log.Printf("UPDATE: Transaction committed and synced for key: %s", string(key))
	return nil
}

// TODO: figure out better abstraction for this, maybe a UpdateWithVersion
func (self *DB) UpdateObj(incomingObj *MetaObject) error {
    key := []byte(incomingObj.ID)

    err := self.db.Update(func(txn *badger.Txn) error {
        // 1. GET current data from DB
        item, err := txn.Get(key)
        
        // --- SCENARIO A: New Item (Create) ---
        if err == badger.ErrKeyNotFound {
            // If the key doesn't exist, we enforce that the client expects Version 0
            if incomingObj.Version != 0 {
                return fmt.Errorf("conflict: item does not exist, but client sent version %d", incomingObj.Version)
            }
            
            // Set first version to 1
            incomingObj.Version = 1
            return saveMetaToBadger(txn, key, incomingObj)
        } else if err != nil {
            return err // legitimate DB error
        }

        // --- SCENARIO B: Update Existing Item ---
        
        // 2. DECODE current data to check version
        valBytes, err := item.ValueCopy(nil)
        if err != nil {
            return err
        }

        var currentObj MetaObject
        if err := json.Unmarshal(valBytes, &currentObj); err != nil {
            return fmt.Errorf("db corruption: failed to unmarshal existing data: %w", err)
        }

        // 3. COMPARE Versions
        // The client says "I am updating Version X".
        // If the DB has Version Y, and X != Y, it's a conflict.
        if incomingObj.Version != currentObj.Version {
            return fmt.Errorf("conflict: client has version %d, but db has version %d", 
                incomingObj.Version, currentObj.Version)
        }

        // 4. INCREMENT and SAVE
        incomingObj.Version = currentObj.Version + 1
        return saveMetaToBadger(txn, key, incomingObj)
    })

    if err != nil {
        return err
    }
    
    return nil
}

func (self *DB) Read(key DBKey) ([]byte, error){
	var retrievedValue []byte
	
	err := self.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		
		if err != nil {
			log.Printf("Error getting key %s: %v", string(key), err)
			return err
		}

		return item.Value(func(val []byte) error {
			retrievedValue = append([]byte{}, val...)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		log.Printf("Key %s not found in database.", string(key))
		return nil, badger.ErrKeyNotFound
	}
	if err != nil {
		log.Printf("Read failed for key %s: %v", string(key), err)
		return nil, err
	}

	return retrievedValue, nil
}

func (self *DB) Delete(key DBKey) error {
	
	err := self.db.Update(func(txn *badger.Txn) error {
			
		log.Printf("DELETE: Attempting to delete key: %s (bytes: %v)", string(key), key)
		err := txn.Delete(key)
		if err != nil {
			log.Printf("DELETE: Error deleting key %s: %v", string(key), err)
			return err
		}
		log.Printf("DELETE: Successfully marked key for deletion: %s", string(key))
        return nil // Returning nil commits the transaction
    })

    if err != nil {
        log.Printf("DELETE: Transaction failed for key %s: %v", string(key), err)
        return err
    }
    
    // 3. Explicitly sync the database to ensure durability on disk
    // Note: You can omit this if you are relying on BadgerDB's default configuration 
    // or are optimizing for performance over immediate durability.
    if err := self.db.Sync(); err != nil {
        log.Printf("DELETE: Sync failed for key %s: %v", string(key), err)
        return err
    }
    
    log.Printf("DELETE: Transaction committed and synced for key: %s", string(key))
    return nil
}


	  ////////////
	 // UTILS  //
	////////////

func saveMetaToBadger(txn *badger.Txn, key []byte, obj *MetaObject) error {
    data, err := json.Marshal(obj)
    if err != nil {
        return fmt.Errorf("failed to marshal: %w", err)
    }
    // We update the entry. 
    return txn.Set(key, data)
}

type DBKey []byte
type Index int
const (
	User Index = iota
	Object
)

func NewDBKey(index Index, id string) DBKey {
	var indexStr string
	switch index {
	case User:
		indexStr = "user"
	case Object:
		indexStr = "objid"
	}
	
	return []byte(fmt.Sprintf("%s:%s", indexStr, id))
}