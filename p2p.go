package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	protocol "github.com/libp2p/go-libp2p-protocol"

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
	tctx, _ := context.WithTimeout(ctx, time.Second*20)
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
					for k, v := c.First(); k != nil; k, v = c.Next() {
						buf := bytes.NewBuffer(v)
						dec := gob.NewDecoder(buf)
						err = dec.Decode(&frq)
						checkError(err)
						go func(key []byte, frq FrReqInd) {
							directlyConnected := false
							connected := false
							pid, err := peer.IDB58Decode(frq.PeerID)
							checkError(err)
							pr, err := dhtClient.FindPeer(tctx, pid)
							if err != nil {
								//host offline.
								//exit the routine
							} else {
								count := 3
								//tries to connect to peer without relay; 3 times.
								for count > 0 {
									if err = thisHost.Connect(tctx, pr); err != nil {
										time.Sleep(time.Second * 2)
									} else {
										directlyConnected = true
										connected = true
									}
								}
								if !directlyConnected {
									//relay via bootstrap node
									relayaddr, err := multiaddr.NewMultiaddr("/p2p-circuit/ipfs/" + pr.ID.Pretty())
									relayPeer := peerstore.PeerInfo{
										ID:    pr.ID,
										Addrs: []multiaddr.Multiaddr{relayaddr},
									}
									count = 3
									for count > 0 {
										if err = thisHost.Connect(tctx, relayPeer); err != nil {
											time.Sleep(time.Second * 2)
										} else {
											connected = true
											break
										}
									}
									if connected {
										//opening stream
										strm, err := thisHost.NewStream(context.Background(), pid, protocolID)
										checkError(err)
										go writeData(key, strm, frq)
									}
								}
							}
						}(k, frq)
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

}

func writeData(key []byte, s net.Stream, frq FrReqInd) {
	//send data
	writer := bufio.NewWriter(bufio.NewWriter(s))
	bytes, err := json.Marshal(frq)
	checkError(err)
	n, err := writer.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil || n == 0 {
		//failed
		fmt.Println("sent")
	} else {
		if err = db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(queueName).Delete(key)
		}); err != nil {
			checkError(err)
		}
	}
}
