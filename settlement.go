package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/segmentio/ksuid"

	"github.com/boltdb/bolt"
)

func settlement(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/settlement.html")
	checkError(err)
	var frd FrdSettle
	var array []ViewSettlement
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(allBkts)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			err = dec.Decode(&frd)
			checkError(err)
			if frd.ID != "" && frd.NickName != "" {
				ows, err := strconv.Atoi(frd.Owes)
				checkError(err)
				own, err := strconv.Atoi(frd.Owns)
				checkError(err)
				msg := ""
				tot := own - ows
				if own <= ows {
					msg = "You owe "
					tot = tot * -1
				} else {
					msg = "Owes you "
				}
				tmp := ViewSettlement{
					NickName: frd.NickName,
					Total:    strconv.Itoa(tot),
					Message:  msg,
				}
				array = append(array, tmp)
			}
		}
		return nil
	})
	if len(array) > 0 {
		resp := struct {
			Found bool
			Data  []ViewSettlement
		}{
			true, array,
		}
		tmpl.Execute(w, resp)
	} else {
		resp := struct {
			Found bool
			Data  []ViewSettlement
		}{
			false, nil,
		}
		tmpl.Execute(w, resp)
	}
}

func viewbill(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == http.MethodGet {
		if uname := r.URL.Query().Get("uname"); uname != "" {
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(uname))
				if b == nil {
					return nil
				}
				c := b.Cursor()
				_ = c
				//to be continued
				return nil
			})
			checkError(err)
		}
	}
}

func addbill(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/addbill.html")
	checkError(err)
	var friends []string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(allBkts)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tmp FrdSettle
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&tmp)
			checkError(err)
			friends = append(friends, tmp.NickName)
		}
		return nil
	})
	if len(friends) > 0 {
		resp := struct {
			Found bool
			Data  []string
		}{
			true, friends,
		}
		tmpl.Execute(w, resp)
	} else {
		resp := struct {
			Found bool
			Data  []string
		}{
			false, nil,
		}
		tmpl.Execute(w, resp)
	}
}

func uploadBill(w http.ResponseWriter, r *http.Request) {
	var friends []string
	var amtSplit []string
	var description string
	var totalAmt string
	var billdt string
	var sentlist []toSentBill
	var uuid string
	var dateNow string
	if r.Method == "POST" {
		dt := time.Now()
		uuid = ksuid.New().String()
		dateNow = dt.Format("02/01/2006")
		err := r.ParseForm()
		checkError(err)
		for key, value := range r.Form {
			if key == "friends[]" {
				friends = value
			} else if key == "amtsplit[]" {
				amtSplit = value
			} else if key == "des" {
				description = value[0]
			} else if key == "tamt" {
				totalAmt = value[0]
			} else if key == "billdt" {
				billdt = value[0]
			}
		}
		if len(friends) != len(amtSplit) || len(friends) == 0 || description == "" || totalAmt == "" || billdt == "" {
			w.Write([]byte("Invalid data provided"))
		} else {
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(friendsBkt)
				if b == nil {
					return nil
				}
				c := b.Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					tmp, err := gobDecodeFrnd(v)
					checkError(err)
					for index, dat := range friends {
						if tmp.NickName == dat {
							stemp := toSentBill{
								id:     tmp.ID,
								name:   tmp.NickName,
								amount: amtSplit[index],
							}
							sentlist = append(sentlist, stemp)
							break
						}
					}
				}
				return nil
			})
			checkError(err)
			var sign SignMe
			var id peer.ID
			if thisHost == nil {
				sKeyName := "prvKey.pem"
				fileFound := true
				file, err := os.Open(sKeyName)
				if err != nil {
					if err.Error() == "open "+sKeyName+": no such file or directory" {
						fileFound = false
					}
				}
				file.Close()
				if !fileFound {
					w.Write([]byte("Create or import credentials first."))
				} else {
					byt, err := ioutil.ReadFile(sKeyName)
					prvKey, err = crypto.UnmarshalRsaPrivateKey(byt)
					if err != nil {

					}
					id, err = peer.IDFromPrivateKey(prvKey)
					checkError(err)
					pubKey = prvKey.GetPublic()
				}
			}
			w.Write([]byte("Bill saved and pushed into queue"))
			go func() {
				var bill BillUpload
				for _, dat := range sentlist {
					sign.UUID = uuid
					sign.DateAdded = dateNow
					sign.Description = description
					sign.Amount = dat.amount
					sign.Date = billdt
					if thisHost == nil {
						sign.HostID = id.Pretty()
					} else {
						sign.HostID = thisHost.ID().Pretty()
					}
					sign.PeerID = dat.id
					var buf bytes.Buffer
					enc := gob.NewEncoder(&buf)
					err = enc.Encode(sign)
					checkError(err)
					signature, err := prvKey.Sign(buf.Bytes())
					checkError(err)
					bill.SignMe = sign
					bill.Signature = signature
					bill.Billup = true
					temp, err := crypto.MarshalPublicKey(pubKey)
					bill.PubKey = temp
					buf.Reset()
					enc = gob.NewEncoder(&buf)
					err = enc.Encode(bill)
					/*
						var sendbill BillUpload
						buf2 := bytes.NewBuffer(buf.Bytes())
						dec := gob.NewDecoder(buf2)
						err = dec.Decode(&sendbill)
						fmt.Println(sendbill)
					*/

					var billInsert BillSave
					billInsert.PeerID = bill.PeerID
					billInsert.Description = bill.Description
					billInsert.Amount = bill.Amount
					billInsert.Date = bill.Date
					billInsert.DateAdded = bill.DateAdded
					var tmpbuf bytes.Buffer
					enc = gob.NewEncoder(&tmpbuf)
					err = enc.Encode(billInsert)
					err = db.Update(func(tx *bolt.Tx) error {
						bucket, err := tx.CreateBucketIfNotExists([]byte(dat.name))
						if err != nil {
							return err
						}
						err = bucket.Put([]byte(sign.UUID), tmpbuf.Bytes())
						return err
					})
					if err == nil {
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
								if tmp.NickName == dat.name {
									//change is necessary here if any user can add anyones bill
									keyToUpdate = k
									dataToUpdate.ID = dat.id
									dataToUpdate.NickName = tmp.NickName
									prev, err := strconv.Atoi(tmp.Owns)
									checkError(err)
									current, err := strconv.Atoi(dat.amount)
									prevOwe, err := strconv.Atoi(tmp.Owes)
									adj := prevOwe - current
									if adj <= 0 {
										dataToUpdate.Owes = "0"
										adj = int(math.Abs(float64(adj)))
										dataToUpdate.Owns = strconv.Itoa(prev + adj)
									} else {
										dataToUpdate.Owes = strconv.Itoa(adj)
										dataToUpdate.Owns = "0"
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
							err = db.Update(func(tx *bolt.Tx) error {
								bucket, err := tx.CreateBucketIfNotExists(queueName)
								if err != nil {
									return err
								}
								tmp, err := bucket.NextSequence()
								checkError(err)
								key := make([]byte, 8)
								binary.LittleEndian.PutUint64(key, uint64(tmp))
								err = bucket.Put(key, buf.Bytes())
								return err
							})
							checkError(err)
						}
					}
				}
			}()
		}
	}
}
