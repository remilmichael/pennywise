package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/boltdb/bolt"
	peer "github.com/libp2p/go-libp2p-peer"
)

func request(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/frq.html")
	checkError(err)
	tmpl.Execute(w, struct{}{})
}

func ajaxreq(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := r.FormValue("download")
		if data == "1" {
			dat, err := boltReturnAll(bucketName)
			checkError(err)
			if len(dat) > 0 {
				resp := struct {
					Empty bool     `json:"Empty"`
					Data  []string `json:"Data"`
				}{
					false, dat,
				}
				jsonrsp, err := json.Marshal(resp)
				checkError(err)
				w.Write(jsonrsp)
			} else {
				resp := struct {
					Empty bool `json:"Empty"`
				}{
					true,
				}
				jsonrsp, err := json.Marshal(resp)
				checkError(err)
				w.Write(jsonrsp)
			}
		} else if data = r.FormValue("name"); data != "" {
			//replace all checkError() function as ajax response
			if thisHost != nil && hostRunning {
				ID, err := returnIDByName([]byte("buddies"), data)
				checkError(err)
				frq := &FrReqInd{
					HostID: thisHost.ID().Pretty(),
					PeerID: ID,
					Flags: Flags{
						FrdAck: false,
						FrdReq: true,
					},
				}
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				err = enc.Encode(frq)

				found, err := queueCompare(queueName, buf.Bytes())
				if found {
					w.Write([]byte("Request already exists in queue."))
				} else {
					err = queueInsert(queueName, buf.Bytes())
					if err == nil {
						w.Write([]byte("Request pushed into the queue."))
					} else {
						w.Write([]byte("Critical error"))
					}
				}
			} else {
				w.Write([]byte("Host is down. Boot host first."))
			}
		}
	}
}

func viewreq(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/viewreq.html")
	checkError(err)
	var dat []string
	if thisHost != nil {
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(reqBkt)
			if b == nil {
				return nil
			}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				dat = append(dat, string(v))
			}
			return err
		})
		if err != nil {
			resp := struct {
				Error   bool
				Success bool
				Data    string
			}{
				true, false, "Fatal error occured",
			}
			tmpl.Execute(w, resp)
		} else {
			if len(dat) > 0 {
				resp := struct {
					Error   bool
					Success bool
					Data    []string
				}{
					false, true, dat,
				}
				tmpl.Execute(w, resp)
			} else {
				resp := struct {
					Error   bool
					Success bool
					Data    []string
				}{
					false, false, []string{},
				}
				tmpl.Execute(w, resp)
			}
		}
	} else {
		resp := struct {
			Error   bool
			Success bool
			Data    string
		}{
			true, false, "Host if offline. Boot host first.",
		}
		tmpl.Execute(w, resp)
	}
}

func processreq(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == "POST" {
		accept := r.FormValue("accept")
		reject := r.FormValue("reject")
		data := r.FormValue("id")
		nickname := r.FormValue("nickname")

		if accept == "1" {
			var frq FrReqInd
			frq.FrdReq = false
			frq.FrdAck = true
			frq.HostID = thisHost.ID().Pretty()
			frq.PeerID = data
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err = enc.Encode(frq)

			_, err = peer.IDB58Decode(data)
			if err != nil {
				w.Write([]byte("Invalid ID"))

			} else {
				frd := &Friend{
					ID:       data,
					NickName: nickname,
				}

				byt, err := gobEncodeFrnd(*frd)
				checkError(err)
				val, err := boltBudSearch(friendsBkt, frd.ID, frd.NickName)
				checkError(err)
				if val == 1 {
					w.Write([]byte("Friend already exists with the ID"))
				} else if val == 2 {
					w.Write([]byte("Nickname already taken. Use another name"))
				} else if val > 2 {
					w.Write([]byte("Can't add same friend twice"))
				} else if val == 0 {
					fmt.Println(frd)
					err = boltInsert(friendsBkt, frd.ID, byt)
					checkError(err)
					//pushing to send queue
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
					w.Write([]byte("Request accepted"))
				}
			}
		} else if reject == "1" {
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(reqBkt)
				if b == nil {
					return nil
				}
				c := b.Cursor()
				tmp := []byte(data)
				for k, v := c.First(); k != nil; k, v = c.Next() {
					if bytes.Equal(v, tmp) {
						if err = db.Update(func(tx *bolt.Tx) error {
							return tx.Bucket(reqBkt).Delete(k)
						}); err != nil {
							checkError(err)
						}
						break
					}
				}
				return err
			})
			checkError(err)
			w.Write([]byte("Request rejected"))
		}
	}
}
