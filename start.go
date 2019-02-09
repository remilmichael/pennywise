package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	mux "github.com/gorilla/mux"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

type output struct {
	Action  bool
	Error   bool
	Success bool
	Msg     string
}

var startOut output
var startTmpl *template.Template

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/start", start)
	r.HandleFunc("/add", addfriend)
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:9090",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func start(w http.ResponseWriter, r *http.Request) {
	var err error
	startTmpl, err = template.ParseFiles("start.html")
	checkError(err)
	genSuccess := false
	if r.Method == http.MethodPost {
		if r.FormValue("create") == "genkey" {
			genSuccess = genKeys()
			if genSuccess {
				flushStartPage(w, true, false, "Keys generated")
			} else {
				flushStartPage(w, false, true, "Key generation failed")
			}
		} else if r.FormValue("import") == "impkey" {
			pubKeySuccess := false
			privKeySuccess := false
			file, _, err := r.FormFile("pubkey")
			if err != nil {
				flushStartPage(w, false, true, "Invalid public key")
			} else {
				defer file.Close()
				fileBytes, err := ioutil.ReadAll(file)
				if err != nil {
					flushStartPage(w, false, true, "Invalid public key")
				} else {
					newFile, err := os.Create("pubKey.pem")
					if err != nil {
						flushStartPage(w, false, true, "Error copying file")
					} else {
						if _, err := newFile.Write(fileBytes); err != nil {
							flushStartPage(w, false, true, "Error copying file")
						} else {
							pubKeySuccess = true
						}
					}
				}
			}
			if pubKeySuccess {
				file, _, err = r.FormFile("privkey")
				if err != nil {
					flushStartPage(w, false, true, "Invalid private key")
				} else {
					defer file.Close()
					fileBytes, err := ioutil.ReadAll(file)
					if err != nil {
						flushStartPage(w, false, true, "Invalid private key")
					} else {
						newFile, err := os.Create("prvKey.pem")
						if err != nil {
							flushStartPage(w, false, true, "Error copying file")
						} else {
							if _, err := newFile.Write(fileBytes); err != nil {
								flushStartPage(w, false, true, "Error copying file")
							} else {
								privKeySuccess = true
							}
						}
					}
				}
			}
			if pubKeySuccess && privKeySuccess {
				flushStartPage(w, true, false, "Key import successful")
			}
		}
	} else {
		flushStartPage(w, false, false, "")
	}
}

func genKeys() bool {
	prvKey, pubKey, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	checkError(err)
	sbyt, err := prvKey.Raw()
	checkError(err)
	pbyt, err := pubKey.Raw()
	checkError(err)
	sfile, err := os.Create("prvKey.pem")
	checkError(err)
	pfile, err := os.Create("pubKey.pem")
	checkError(err)
	_, err = sfile.Write(sbyt)
	checkError(err)
	_, err = pfile.Write(pbyt)
	checkError(err)
	return true
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func flushStartPage(w http.ResponseWriter, e bool, act bool, msg string) {
	startOut.Action = act
	startOut.Error = e
	startOut.Msg = msg
	startTmpl.Execute(w, startOut)
}
