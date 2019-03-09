package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
)

func processData(str string) {
	var data Receive
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		checkError(err)
	}
	if data.Flags.FrdReq {
		saveFrdReq(str)
	} else if data.Flags.FrdAck {
		//to be continued
	}
}

func saveFrdReq(str string) {
	var err error
	var frq FrReqInd
	if err = json.Unmarshal([]byte(str), &frq); err != nil {
		//ignore
		log.Println(err)
	} else {
		hid := []byte(frq.HostID)
		found := false
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(reqBkt)
			if b == nil {
				return nil
			}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if bytes.Equal(v, hid) {
					found = true
					break
				}
			}
			return err
		})
		if !found {
			err = db.Update(func(tx *bolt.Tx) error {
				bucket, err := tx.CreateBucketIfNotExists(reqBkt)
				if err != nil {
					return err
				}
				tmp, err := bucket.NextSequence()
				checkError(err)
				key := make([]byte, 8)
				binary.LittleEndian.PutUint64(key, uint64(tmp))
				err = bucket.Put(key, []byte(frq.HostID))
				return err
			})
			checkError(err)
		}
	}
}
