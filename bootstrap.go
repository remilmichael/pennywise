package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	crypto "github.com/libp2p/go-libp2p-crypto"
)

var pubKey crypto.PubKey
var prvKey crypto.PrivKey

func bootstrap(w http.ResponseWriter, r *http.Request) {
	type redir struct {
		Redirect bool
	}
	var out redir
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
	}
}
