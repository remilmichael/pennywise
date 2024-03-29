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
	houseKeeping(1)
	hostRunning = false
	r := mux.NewRouter()
	r.HandleFunc("/start", start)
	r.HandleFunc("/add", addfriend)
	r.HandleFunc("/boot", bootstrap)
	r.HandleFunc("/request", request)
	r.HandleFunc("/ajaxreq", ajaxreq)
	r.HandleFunc("/viewreq", viewreq)
	r.HandleFunc("/processreq", processreq)
	r.HandleFunc("/settlement", settlement)
	r.HandleFunc("/viewbill", viewbill)
	r.HandleFunc("/addbill", addbill)
	r.HandleFunc("/uploadbill", uploadBill)
	r.HandleFunc("/", home)
	r.HandleFunc("/index", indexpage)

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
	startTmpl, err = template.ParseFiles("html/start.html")
	checkError(err)
	genSuccess := false
	if r.Method == http.MethodPost {
		if r.FormValue("create") == "genkey" {
			genSuccess = genKeys()
			if genSuccess {
				flushStartPage(w, true, true, false, "Keys generated")
			} else {
				flushStartPage(w, true, false, true, "Key generation failed")
			}
		} else if r.FormValue("import") == "impkey" {
			pubKeySuccess := false
			privKeySuccess := false
			file, _, err := r.FormFile("pubKey")
			if err != nil {
				flushStartPage(w, true, false, true, "Invalid public key")
			} else {
				defer file.Close()
				fileBytes, err := ioutil.ReadAll(file)
				if err != nil {
					flushStartPage(w, true, false, true, "Invalid public key")
				} else {
					newFile, err := os.Create("pubKey.pem")
					if err != nil {
						flushStartPage(w, true, false, true, "Error copying file")
					} else {
						if _, err := newFile.Write(fileBytes); err != nil {
							flushStartPage(w, true, false, true, "Error copying file")
						} else {
							pubKeySuccess = true
						}
					}
				}
			}
			if pubKeySuccess {
				file, _, err = r.FormFile("prvKey")
				if err != nil {
					flushStartPage(w, true, false, true, "Invalid private key")
				} else {
					defer file.Close()
					fileBytes, err := ioutil.ReadAll(file)
					if err != nil {
						flushStartPage(w, true, false, true, "Invalid private key")
					} else {
						newFile, err := os.Create("prvKey.pem")
						if err != nil {
							flushStartPage(w, true, false, true, "Error copying file")
						} else {
							if _, err := newFile.Write(fileBytes); err != nil {
								flushStartPage(w, true, false, true, "Error copying file")
							} else {
								privKeySuccess = true
							}
						}
					}
				}
			}
			if pubKeySuccess && privKeySuccess {
				flushStartPage(w, true, true, false, "Key import successful")
			}
		}
	} else {
		flushStartPage(w, false, false, false, "")
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
		log.Panic(err)
	}
}

func flushStartPage(w http.ResponseWriter, act bool, s bool, e bool, msg string) {
	startOut.Action = act
	startOut.Success = s
	startOut.Error = e
	startOut.Msg = msg
	startTmpl.Execute(w, startOut)
}

func houseKeeping(state int8) {
	var err error
	if state == 1 {
		err = boltOpen()
		if err != nil {
			checkError(err)
		}

	}
}

func home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/home.html")
	checkError(err)
	tmpl.Execute(w, "")
}

func indexpage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/index.html")
	checkError(err)
	tmpl.Execute(w, "")
}
