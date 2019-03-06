package main

import (
	"bufio"
	"bytes"
	"context"
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
				if recv.FrdReq == true {
					var frq FrReqInd
					err = dec.Decode(&frq)
					pid, err = peer.IDB58Decode(frq.PeerID)
					checkError(err)
					byteData, err = json.Marshal(frq)
					checkError(err)
				} else if recv.FrdAck == true {
					var frq FrReqInd
					err = dec.Decode(&frq)
					buf = bytes.NewBuffer(v)
					pid, err = peer.IDB58Decode(frq.HostID)
					checkError(err)
					byteData, err = json.Marshal(frq)
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
						writeData(rw, s, pid, byteData, key)
						s.Close()
						thisHost.Network().ClosePeer(pid)
					}
				}
			}
			time.Sleep(time.Second * 5)
		}
		return nil
	})
	checkError(err)
}
