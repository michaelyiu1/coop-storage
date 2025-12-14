package main
import (
	"log"

)
// In this module: logic to interface between the client and the metadata server

// server -> client payload
type OSDGuide struct {
	Token string `json:"Token"`
	PageNum int `json:"PageNum"` //TODO: implement paging
	IDMap map[string]string `json:"IDMap"`
}

func (self *OSDGuide) Read(user string) (error) {
	self.IDMap = make(map[string]string)
	
	for o, err := range IDNameTupleIterator(user) {
		if err != nil {
			log.Printf("error, %v", err)
			return err
		}
		if o.FileName == "" {
			return nil
		}
		self.IDMap[o.FileName] = o.ID
	}
	return nil
}
