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

	protocol "github.com/libp2p/go-libp2p-protocol"
	swarm "github.com/libp2p/go-libp2p-swarm"

	"github.com/boltdb/bolt"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func sendReq() {
	var protocolID protocol.ID = "/pwise"
	var err error
	var frq FrReqInd
	ctx := context.Background()
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	thisHost.SetStreamHandler(protocolID, handleStream)
	for {
		if thisHost != nil && hostRunning {
			for {
				err = db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket(queueName)
					if b == nil {
						return nil
					}
					c := b.Cursor()
					for key, v := c.First(); key != nil; key, v = c.Next() {
						buf := bytes.NewBuffer(v)
						dec := gob.NewDecoder(buf)
						err = dec.Decode(&frq)
						fmt.Println(frq)
						checkError(err)
						directlyConnected := false
						connected := false
						pid, err := peer.IDB58Decode(frq.PeerID)
						checkError(err)
						tctx, cancel = context.WithTimeout(ctx, time.Second*10)
						pr, err := dhtClient.FindPeer(tctx, pid)
						cancel()
						fmt.Println(pr)
						if err != nil {
							//host offline.
							//exit the routine
						} else {
							count := 3
							//tries to connect to peer without relay; 3 times.
							for count > 0 {
								fmt.Println("trying without relay")
								tctx, cancel = context.WithTimeout(ctx, time.Second*10)
								if err = thisHost.Connect(tctx, pr); err != nil {
									thisHost.Network().(*swarm.Swarm).Backoff().Clear(pr.ID)
									fmt.Println(err)
									time.Sleep(time.Second * 2)
								} else {
									directlyConnected = true
									connected = true
									break
								}
								count--
							}
							cancel()
							if !directlyConnected {
								//relay via bootstrap node
								relayaddr, err := multiaddr.NewMultiaddr("/p2p-circuit/ipfs/" + pid.Pretty())
								relayPeer := peerstore.PeerInfo{
									ID:    pid,
									Addrs: []multiaddr.Multiaddr{relayaddr},
								}
								count = 3
								for count > 0 {
									fmt.Println("trying with relay")
									if err = thisHost.Connect(context.Background(), relayPeer); err != nil {
										thisHost.Network().(*swarm.Swarm).Backoff().Clear(relayPeer.ID)
										time.Sleep(time.Second * 2)
										fmt.Println("directly", err)
									} else {
										connected = true
										break
									}
									count--
								}
							}
							if connected {
								fmt.Println("yoyo")
								//opening stream
								strm, err := thisHost.NewStream(context.Background(), pid, protocolID)
								checkError(err)
								writeData(key, strm, frq)
							}
						}
					}
					return nil
				})
				time.Sleep(time.Second * 10)
			}
		}
		time.Sleep(time.Second * 15)
	}
}

func handleStream(s net.Stream) {
	go readData(s)
}

func writeData(key []byte, s net.Stream, frq FrReqInd) {
	//send data
	fmt.Println("yo")
	writer := bufio.NewWriter(bufio.NewWriter(s))
	bytes, err := json.Marshal(frq)
	checkError(err)
	n, err := writer.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil || n == 0 {
		//failed
		fmt.Println("send error", err)
	} else {
		fmt.Println("sent")
		if err = db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(queueName).Delete(key)
		}); err != nil {
			checkError(err)
		}
	}
	time.Sleep(time.Second * 2)
	s.Close()
}

func readData(s net.Stream) {
	reader := bufio.NewReader(bufio.NewReader(s))
	str, err := reader.ReadString('\n')
	if err != nil {
		if err.Error() == "stream reset" {
			//
		} else {
			log.Panic(err)
		}
	}
	var rcvd FrReqInd
	if err := json.Unmarshal([]byte(str), &rcvd); err != nil {
		checkError(err)
	}
	fmt.Println(rcvd)
}
