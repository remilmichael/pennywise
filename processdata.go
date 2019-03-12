package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
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
		saveFrdAck(str)
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

//save friend to db
func saveFrdAck(str string) {
	var frq FrReqInd
	var err error
	var name string
	var id string
	var alreadyFriend bool
	if err = json.Unmarshal([]byte(str), &frq); err != nil {
		log.Println(err)
	} else {
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(friendsBkt)
			if b == nil {
				return nil
			}
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				if string(k) == frq.PeerID {
					alreadyFriend = true
					break
				}
			}
			return nil
		})
		checkError(err)
		if !alreadyFriend {
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(buddyBkt)
				if b == nil {
					return nil
				}
				c := b.Cursor()
				var tmp *Friend
				for k, v := c.First(); k != nil; k, v = c.Next() {
					tmp, err = gobDecodeFrnd(v)
					if err != nil {
						return err
					}
					if tmp.ID == frq.HostID {
						name = tmp.NickName
						id = tmp.ID
						break
					}
				}
				return nil
			})
			if name != "" {
				frd := &Friend{
					ID:       id,
					NickName: name,
				}
				byt, err := gobEncodeFrnd(*frd)
				checkError(err)
				err = db.Update(func(tx *bolt.Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(friendsBkt)
					if err != nil {
						return err
					}
					err = bucket.Put([]byte(frd.ID), byt)
					return nil
				})
				checkError(err)
				frdtmp := &FrdSettle{
					ID:       id,
					NickName: name,
					Total:    "0",
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				err = enc.Encode(frdtmp)
				checkError(err)
				err = db.Update(func(tx *bolt.Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(allBkts)
					if err != nil {
						return err
					}
					tmp, err := bucket.NextSequence()
					if err != nil {
						checkError(err)
					}
					key := make([]byte, 8)
					binary.LittleEndian.PutUint64(key, uint64(tmp))
					err = bucket.Put(key, buf.Bytes())
					return nil
				})
				checkError(err)
			}
		}
	}
}
