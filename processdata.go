package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"log"
	"math"
	"strconv"

	"github.com/boltdb/bolt"
	crypto "github.com/libp2p/go-libp2p-crypto"
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
	} else if data.Flags.Billup {
		saveBill(str)
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
					Owns:     "0",
					Owes:     "0",
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

func saveBill(str string) {
	var err error
	var bill BillUpload
	var sigbuf bytes.Buffer
	if err = json.Unmarshal([]byte(str), &bill); err != nil {
		//ignore
		log.Println(err)
	} else {
		var verifyme SignMe
		verifyme.UUID = bill.UUID
		verifyme.HostID = bill.HostID
		verifyme.PeerID = bill.PeerID
		verifyme.Description = bill.Description
		verifyme.Amount = bill.Amount
		verifyme.Date = bill.Date
		verifyme.DateAdded = bill.DateAdded
		out := make(chan bool)
		go func(out chan<- bool) {
			sigbuf.Reset()
			encod := gob.NewEncoder(&sigbuf)
			err = encod.Encode(verifyme)
			out <- true
		}(out)
		<-out
		checkError(err)
		peerPubkey, err := crypto.UnmarshalRsaPublicKey(bill.PubKey)
		checkError(err)
		signValid, err := peerPubkey.Verify(sigbuf.Bytes(), bill.Signature)
		if signValid {
			var frd Friend
			var name string
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(friendsBkt)
				if b == nil {
					return nil
				}
				c := b.Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					buf := bytes.NewBuffer(v)
					dec := gob.NewDecoder(buf)
					err = dec.Decode(&frd)
					if frd.ID == verifyme.HostID {
						name = frd.NickName
						break
					}
				}
				return nil
			})
			checkError(err)
			if name != "" {
				var billInsert BillSave
				billInsert.PeerID = verifyme.HostID
				billInsert.Description = verifyme.Description
				billInsert.Amount = verifyme.Amount
				billInsert.Date = verifyme.Date
				billInsert.DateAdded = verifyme.DateAdded
				billInsert.Type = 2
				var insertbuf bytes.Buffer
				enc := gob.NewEncoder(&insertbuf)
				err = enc.Encode(billInsert)
				checkError(err)
				err = db.Update(func(tx *bolt.Tx) error {
					bucket, err := tx.CreateBucketIfNotExists([]byte(name))
					if err != nil {
						return err
					}
					key := []byte(verifyme.UUID)
					err = bucket.Put(key, insertbuf.Bytes())
					return err
				})
				checkError(err)
				var gotData bool
				var keyToUpdate []byte
				var dataToUpdate FrdSettle
				err = db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket(allBkts)
					if b == nil {
						return nil
					}
					c := b.Cursor()
					for k, v := c.First(); k != nil; k, v = c.Next() {
						var tmp FrdSettle
						buf2 := bytes.NewBuffer(v)
						dec := gob.NewDecoder(buf2)
						err = dec.Decode(&tmp)
						if tmp.ID == verifyme.HostID {
							keyToUpdate = k
							dataToUpdate.ID = tmp.ID
							dataToUpdate.NickName = tmp.NickName
							oweprev, err := strconv.Atoi(tmp.Owes)
							checkError(err)
							owecurr, err := strconv.Atoi(verifyme.Amount)
							checkError(err)
							prevOwn, err := strconv.Atoi(tmp.Owns)
							adj := prevOwn - owecurr
							if adj <= 0 {
								dataToUpdate.Owns = "0"
								adj = int(math.Abs(float64(adj)))
								dataToUpdate.Owes = strconv.Itoa(oweprev + adj)
							} else {
								dataToUpdate.Owns = strconv.Itoa(adj)
								dataToUpdate.Owes = "0"
							}
							gotData = true
							break
						}
					}
					return nil
				})
				checkError(err)
				if gotData {
					err = db.Update(func(tx *bolt.Tx) error {
						bucket, err := tx.CreateBucketIfNotExists(allBkts)
						if err != nil {
							return err
						}
						var buftemp bytes.Buffer
						enc := gob.NewEncoder(&buftemp)
						err = enc.Encode(dataToUpdate)
						err = bucket.Put(keyToUpdate, buftemp.Bytes())
						return err
					})
				}
			}
		} else {
			log.Println(err)
		}
	}
}
