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

func writeData(rw *bufio.ReadWriter, s net.Stream, pid peer.ID, bytes []byte, key []byte) {
	_, err := rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil {
		if err.Error() == "stream reset" {
			s.Close()
			thisHost.Network().ClosePeer(pid)
			return
		}
	}
	rw.Flush()
	time.Sleep(time.Second * 1)
	/*if err = db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(queueName).Delete(key)
	}); err != nil {
		checkError(err)
	}*/
}

func sendReq() {
	var frq FrReqInd
	var pid peer.ID
	var byteData []byte
	ctx := context.Background()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(queueName)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for {
			for key, v := c.First(); key != nil; key, v = c.Next() {
				var recv Receive
				buf := bytes.NewBuffer(v)
				dec := gob.NewDecoder(buf)
				err := dec.Decode(&recv)

				/*var frq FrReqInd
				buf = bytes.NewBuffer(v)
				dec = gob.NewDecoder(buf)
				err = dec.Decode(&frq)
				fmt.Println(frq)*/

				if recv.FrdReq == true || recv.FrdAck == true {
					buf = bytes.NewBuffer(v)
					dec = gob.NewDecoder(buf)
					err = dec.Decode(&frq)
					pid, err = peer.IDB58Decode(frq.PeerID)
					checkError(err)
					byteData, err = json.Marshal(frq)
					checkError(err)
				}
				tctx, _ := context.WithTimeout(ctx, time.Second*10)
				pr, err := dhtClient.FindPeer(tctx, pid)
				if err != nil {
					//ignore
				} else {
					fmt.Println(pr)
					if err = thisHost.Connect(tctx, pr); err != nil {
						thisHost.Network().(*swarm.Swarm).Backoff().Clear(pr.ID)
					} else {
						fmt.Println("connected")
						s, err := thisHost.NewStream(context.Background(), pid, "/cats")
						if err != nil {
							log.Println(err)
							continue
						}
						rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
						//time.Sleep(time.Second * 2)
						writeData(rw, s, pid, byteData, key)
						s.Close()
						thisHost.Network().ClosePeer(pid)
						if recv.FrdReq == true {
							found := false
							err = db.View(func(tx *bolt.Tx) error {
								b := tx.Bucket(reqdump)
								if b == nil {
									return nil
								}
								c := b.Cursor()
								tmp := []byte(frq.PeerID)
								for k, v := c.First(); k != nil; k, v = c.Next() {
									if bytes.Equal(v, tmp) {
										found = true
										break
									}
								}
								return err
							})
							if !found {
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
						}
					}
				}
			}
			time.Sleep(time.Second * 5)
		}
		return nil
	})
	checkError(err)
}
