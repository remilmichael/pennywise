package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	levelds "github.com/ipfs/go-ds-leveldb"
	ipns "github.com/ipfs/go-ipns"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func main() {
	fileName := "keys/dht_prvKey.pem"
	ctx := context.Background()
	var byt []byte
	var prvKey crypto.PrivKey
	fileFound := true
	file, err := os.Open(fileName)
	if err != nil {
		if err.Error() == "open prvKey.pem: no such file or directory" {
			fileFound = false
		}
	}
	file.Close()
	if !fileFound {
		prvKey, _, err = crypto.GenerateKeyPair(crypto.RSA, 2048)
		checkError(err)
		byt, _ := prvKey.Raw()
		file, err = os.Create(fileName)
		checkError(err)
		_, err = file.Write(byt)
		if err != nil {
			panic(err)
		}
		file.Close()
	} else {
		byt, err = ioutil.ReadFile(fileName)
		checkError(err)
		prvKey, err = crypto.UnmarshalRsaPrivateKey(byt)
		checkError(err)
	}

	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
		libp2p.Identity(prvKey),
		libp2p.NATPortMap(),
	)
	checkError(err)
	/*
		fmt.Println("Address = ", host.Addrs())
		fmt.Println("Host ID = ", host.ID().Pretty())
	*/
	temp := host.Addrs()
	var addr multiaddr.Multiaddr
	for _, i := range temp {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	hAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID().Pretty()))
	fullAddr := addr.Encapsulate(hAddr)
	fmt.Println("Full address = ", fullAddr)

	ds, err := levelds.NewDatastore("dbase", nil)
	checkError(err)
	d := dht.NewDHT(ctx, host, ds)
	d.Validator = record.NamespacedValidator{
		"pk":   record.PublicKeyValidator{},
		"ipns": ipns.Validator{KeyBook: host.Peerstore()},
	}

	err = d.Bootstrap(ctx)
	if err != nil {
		log.Println(err)
	}

	//To check if peers are connected
	/*
		go func() {
			fmt.Println("\n==List of available peers==")
			for {
				connPeers := host.Network().Peers()
				if len(connPeers) > 0 {
					for _, p := range connPeers {
						fmt.Println(p, " = ", host.Network().Connectedness(p))
					}
				}
				time.Sleep(time.Second * 5)
			}
		}()
	*/
	select {}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
