package main

import (
	"log"
	"github.com/dgraph-io/badger/v4"
)


type DB struct{
	db *badger.DB
}

var DBInst *DB

func InitDb() {

	opts := badger.DefaultOptions(DB_PATH)

	// opts.Logger = nil

	db, err := badger.Open(opts)
	DBInst = &DB {db}

	if err != nil {
		log.Fatalf("Error opening BadgerDB: %v", err)
	}

}

func CloseDb() {
	if err := DBInst.db.Close(); err != nil {
		log.Printf("Error closing BadgerDB: %v", err)
	}
}

func (self *DB) Update(key []byte, val []byte) (error) {
	log.Printf("UPDATE: Writing key: %s (bytes: %v), value length: %d", string(key), key, len(val))

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

func (self *DB) Read(key []byte) ([]byte, error){
	var retrievedValue []byte

	log.Printf("READ: Attempting to read key: %s (bytes: %v)", string(key), key)
	
	err := self.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		
		if err != nil {
			log.Printf("Error getting key %s: %v", string(key), err)
			return err
		}

		return item.Value(func(val []byte) error {
			retrievedValue = append([]byte{}, val...)
			log.Printf("READ: Successfully read key: %s, value length: %d", string(key), len(retrievedValue))
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

func (self *DB) Delete(key []byte) error {
    log.Printf("DELETE: Attempting to delete key: %s (bytes: %v)", string(key), key)

    // 1. Start an Update transaction
    err := self.db.Update(func(txn *badger.Txn) error {
        // 2. Perform the delete operation within the transaction
        err := txn.Delete(key)
        if err != nil {
            // Log the error if the deletion fails (e.g., I/O error, not necessarily key not found)
            log.Printf("DELETE: Error deleting key %s: %v", string(key), err)
            return err
        }
        log.Printf("DELETE: Successfully marked key for deletion: %s", string(key))
        return nil // Returning nil commits the transaction
    })

    // Handle transaction commit failure (e.g., conflict, write error)
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