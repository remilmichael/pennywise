package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
)

func handleStream(s net.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(s, rw)
}

func readData(s net.Stream, rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			if err.Error() == "stream reset" || err.Error() == "EOF" {
				break
			} else {
				log.Println(err)
			}
		} else {
			go processData(str)
		}
		time.Sleep(time.Second * 1)
	}
}

func writeData(rw *bufio.ReadWriter, s net.Stream, pid peer.ID, bytes []byte) bool {
	_, err := rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil {
		if err.Error() == "stream reset" {
			s.Close()
			thisHost.Network().ClosePeer(pid)
			return false
		}
	}
	rw.Flush()
	time.Sleep(time.Second * 1)
	return true
}

func sendReq() {
	var frq FrReqInd
	var pid peer.ID
	var sendbill BillUpload
	var byteData []byte
	var keyToDelete []byte
	var transmitDone bool
	ctx := context.Background()
	foundReqdump := false
	for {
		transmitDone = false
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(queueName)
			if b == nil {
				return nil
			}
			c := b.Cursor()
			for key, v := c.First(); key != nil; key, v = c.Next() {
				keyToDelete = nil
				transmitDone = false
				var recv Receive
				buf := bytes.NewBuffer(v)
				dec := gob.NewDecoder(buf)
				err := dec.Decode(&recv)
				if recv.FrdReq == true || recv.FrdAck == true {
					buf = bytes.NewBuffer(v)
					dec = gob.NewDecoder(buf)
					err = dec.Decode(&frq)
					pid, err = peer.IDB58Decode(frq.PeerID)
					checkError(err)
					byteData, err = json.Marshal(frq)
					checkError(err)
				} else if recv.Billup {
					buf = bytes.NewBuffer(v)
					dec = gob.NewDecoder(buf)
					err = dec.Decode(&sendbill)
					checkError(err)
					pid, err = peer.IDB58Decode(sendbill.SignMe.PeerID)
					checkError(err)
					byteData, err = json.Marshal(sendbill)
					checkError(err)
				}
				tctx, _ := context.WithTimeout(ctx, time.Second*10)
				pr, err := dhtClient.FindPeer(tctx, pid)
				if err != nil {
					//ignore
				} else {
					if err = thisHost.Connect(tctx, pr); err != nil {
						thisHost.Network().(*swarm.Swarm).Backoff().Clear(pr.ID)
					} else {
						s, err := thisHost.NewStream(context.Background(), pid, "/cats")
						if err != nil {
							log.Println(err)
							continue
						}
						rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
						//time.Sleep(time.Second * 2)
						if writeData(rw, s, pid, byteData) {
							keyToDelete = key
						}
						s.Close()
						thisHost.Network().ClosePeer(pid)
						transmitDone = true
						if recv.FrdReq == true {
							foundReqdump = false
							err = db.View(func(tx *bolt.Tx) error {
								b := tx.Bucket(reqdump)
								if b == nil {
									return nil
								}
								c := b.Cursor()
								tmp := []byte(frq.PeerID)
								for k, v := c.First(); k != nil; k, v = c.Next() {
									if bytes.Equal(v, tmp) {
										foundReqdump = true
										break
									}
								}
								return err
							})
						}
					}
				}
			}
			return nil
		})
		checkError(err)
		if !foundReqdump && transmitDone {
			err = db.Update(func(tx *bolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists(reqdump)
				if err != nil {
					return err
				}
				tmp, err := b.NextSequence()
				if err != nil {
					checkError(err)
				}
				key := make([]byte, 8)
				binary.LittleEndian.PutUint64(key, uint64(tmp))
				err = b.Put(key, []byte(frq.PeerID))
				return err
			})
			checkError(err)
		}

		if keyToDelete != nil {
			if err = db.Update(func(tx *bolt.Tx) error {
				return tx.Bucket(queueName).Delete(keyToDelete)
			}); err != nil {
				log.Println(err)
			}
		}
		time.Sleep(time.Second * 5)
	}
}
