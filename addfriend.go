package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"

	peer "github.com/libp2p/go-libp2p-peer"
)

type Friend struct {
	ID       string
	NickName string
}

var out output
var addFrdTmpl *template.Template

func addfriend(w http.ResponseWriter, r *http.Request) {
	var bucketName = "buddies"
	var err error
	var nickName string
	var budID string
	//var pid peer.ID
	idOkay := false
	nameOkay := false
	addFrdTmpl, err = template.ParseFiles("addfriend.html")
	checkError(err)
	if r.Method == http.MethodPost {
		if r.FormValue("save") == "savetodb" {
			budID = r.FormValue("budid")
			_, err = peer.IDB58Decode(budID)
			if err != nil {
				flushAddFrdPage(w, true, true, false, "Invalid ID provided!")
			} else {
				idOkay = true
			}
			nickName = r.FormValue("nickname")
			if idOkay {
				if len(nickName) < 1 {
					flushAddFrdPage(w, true, true, false, "No nickname provided")
				} else {
					nameOkay = true
				}
			}
		}
		if idOkay && nameOkay {
			frd := &Friend{
				ID:       budID,
				NickName: nickName,
			}
			byt, err := gobEncodeFrnd(*frd)
			_ = byt
			if err == nil {
				replace := false
				replaceVal := ""
				err = boltOpen()
				if err != nil {
					panic(err)
				}
				defer boltClose()
				rbyt, err := boltSelect(bucketName, budID)
				_ = rbyt
				if err != nil {
					panic(err)
				}
				if len(rbyt) > 0 {
					da, err := gobDecodeFrnd(rbyt)
					if err != nil {
						panic(err)
					}
					replace = true
					replaceVal = da.NickName
				}
				err = boltInsert(bucketName, frd.ID, byt)
				if err != nil {
					flushAddFrdPage(w, true, true, false, "Error saving data")
				} else {
					if replace {
						flushAddFrdPage(w, true, false, true, fmt.Sprintf("Buddy name replaced : %s with %s", replaceVal, nickName))
					} else {
						flushAddFrdPage(w, true, false, true, "Buddy added")
					}
				}
			} else {
				flushAddFrdPage(w, true, true, false, "Error saving data")
			}
		}
	} else {
		flushAddFrdPage(w, false, false, false, "")
	}
}

func flushAddFrdPage(w http.ResponseWriter, act bool, e bool, s bool, msg string) {
	out.Action = act
	out.Error = e
	out.Success = s
	out.Msg = msg
	addFrdTmpl.Execute(w, out)
}

func gobEncodeFrnd(fr Friend) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(fr)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecodeFrnd(byt []byte) (*Friend, error) {
	var frd *Friend
	buf := bytes.NewBuffer(byt)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&frd)
	if err != nil {
		return frd, err
	}
	return frd, nil
}
