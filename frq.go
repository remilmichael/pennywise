package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"html/template"
	"net/http"
)

var queueName []byte

func request(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/frq.html")
	checkError(err)
	tmpl.Execute(w, struct{}{})
}

func ajaxreq(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data := r.FormValue("download")
		if data == "1" {
			dat, err := boltReturnAll([]byte("buddies"))
			checkError(err)
			if len(dat) > 1 {
				resp := struct {
					Empty bool     `json:"Empty"`
					Data  []string `json:"Data"`
				}{
					false, dat,
				}
				jsonrsp, err := json.Marshal(resp)
				checkError(err)
				w.Write(jsonrsp)
			} else {
				resp := struct {
					Empty bool `json:"Empty"`
				}{
					true,
				}
				jsonrsp, err := json.Marshal(resp)
				checkError(err)
				w.Write(jsonrsp)
			}
		} else if data = r.FormValue("name"); data != "" {
			//replace all checkError() function as ajax response
			if thisHost != nil && hostRunning {
				ID, err := returnIDByName([]byte("buddies"), data)
				checkError(err)
				var frq FrReqInd
				frq.flag.FrdReq = true
				frq.HostID = thisHost.ID().Pretty()
				frq.PeerID = ID
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				err = enc.Encode(frq)
				found, err := queueCompare(queueName, buf.Bytes())
				if found {
					w.Write([]byte("Request already exists in queue."))
				} else {
					err = queueInsert(queueName, buf.Bytes())
					if err != nil {
						w.Write([]byte("Request pushed into the queue."))
					}
				}
			} else {
				w.Write([]byte("Host is down. Boot host first."))
			}
		}
	}
}
