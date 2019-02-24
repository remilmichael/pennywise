package main

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multihash "github.com/multiformats/go-multihash"

	cid "github.com/ipfs/go-cid"
	datastore "github.com/ipfs/go-datastore"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	libp2p "github.com/libp2p/go-libp2p"

	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
)

var pubKey crypto.PubKey
var prvKey crypto.PrivKey
var thisHost host.Host
var dhtClient *dht.IpfsDHT
var bootStrapPeer string
var rendezvous string
var hostRunning bool

func bootstrap(w http.ResponseWriter, r *http.Request) {
	rendezvous = "pennywise"
	ctx := context.Background()
	bootStrapPeer = "/ip4/13.58.140.223/tcp/4001/ipfs/QmTAjBx9QfmPqMKTrAG7tfPEEJDtH7oFhecz3TqJgjppk1"
	if !hostRunning {
		type BootPage struct {
			Redirect bool
			Error    bool
			Msg      string
		}
		var out BootPage
		tmpl, err := template.ParseFiles("html/boot.html")
		if err != nil {
			panic(err)
		}
		sKeyName := "prvKey.pem"
		fileFound := true
		file, err := os.Open(sKeyName)
		if err != nil {
			if err.Error() == "open prvKey.pem: no such file or directory" {
				fileFound = false
			}
		}
		file.Close()
		if !fileFound {
			out.Redirect = true
			tmpl.Execute(w, out)
		} else {
			byt, err := ioutil.ReadFile(sKeyName)
			checkError(err)
			prvKey, err = crypto.UnmarshalRsaPrivateKey(byt)
			if err != nil {
				out.Redirect = true
				tmpl.Execute(w, out)
			}
			thisHost, err = libp2p.New(
				ctx,
				libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001", "/ip6/::/tcp/4001"),
				libp2p.Identity(prvKey),
				libp2p.NATPortMap(),
			)
			if err != nil {
				out.Error = true
				out.Msg = "Unable to boot host"
				tmpl.Execute(w, out)
			} else {
				dhtClient = dht.NewDHTClient(ctx, thisHost, datastore.NewMapDatastore())
				bootAddr, _ := ipfsaddr.ParseString(bootStrapPeer)
				bootInfo, _ := peerstore.InfoFromP2pAddr(bootAddr.Multiaddr())
				outChan := make(chan bool)
				go func(outChan chan<- bool) {
					bootCount := 0
					for {
						if bootCount > 5 {
							outChan <- false
							break
						}
						if err = thisHost.Connect(ctx, *bootInfo); err != nil {
							time.Sleep(time.Second * 5)
						} else {
							outChan <- true
							break
						}
						bootCount++
					}
				}(outChan)
				if !<-outChan {
					thisHost.Close()
					hostRunning = false
					out.Error = true
					out.Msg = "Bootstrapping failed."
					tmpl.Execute(w, out)
				} else {
					hostRunning = true
					pref := cid.V1Builder{
						Codec:  cid.Raw,
						MhType: multihash.SHA2_256,
					}
					contID, err := pref.Sum([]byte(rendezvous))
					checkError(err)
					tctx, ctxCancel := context.WithTimeout(ctx, time.Second*10)
					_ = ctxCancel

					//Announcing
					go func() {
						for {
							if err = dhtClient.Provide(tctx, contID, true); err != nil {
								time.Sleep(time.Second * 3)
							} else {
								break
							}
						}
					}()
					if hostRunning {
						go sendReq()
					}
				}
			}
		}
	}

}
