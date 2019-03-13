package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"

	"github.com/boltdb/bolt"
)

func settlement(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/settlement.html")
	checkError(err)
	var friends []string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(allBkts)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.First() {
			friends = append(friends, string(v))
		}
		return nil
	})
	if len(friends) > 0 {
		resp := struct {
			Found bool
			Data  []string
		}{
			true, friends,
		}
		tmpl.Execute(w, resp)
	} else {
		resp := struct {
			Found bool
			Data  []string
		}{
			false, nil,
		}
		tmpl.Execute(w, resp)
	}
}

func viewbill(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == http.MethodGet {
		if uname := r.URL.Query().Get("uname"); uname != "" {
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(uname))
				if b == nil {
					return nil
				}
				c := b.Cursor()
				_ = c
				//to be continued
				return nil
			})
			checkError(err)
		}
	}
}

func addbill(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/addbill.html")
	checkError(err)
	var friends []string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(allBkts)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tmp FrdSettle
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&tmp)
			checkError(err)
			friends = append(friends, tmp.NickName)
		}
		return nil
	})
	if len(friends) > 0 {
		resp := struct {
			Found bool
			Data  []string
		}{
			true, friends,
		}
		tmpl.Execute(w, resp)
	} else {
		resp := struct {
			Found bool
			Data  []string
		}{
			false, nil,
		}
		tmpl.Execute(w, resp)
	}
}

func uploadBill(w http.ResponseWriter, r *http.Request) {
	var friends []string
	var amtSplit []string
	var descrption string
	var totalAmt string
	var billdt string
	var sentlist []toSentBill
	if r.Method == "POST" {
		err := r.ParseForm()
		checkError(err)
		for key, value := range r.Form {
			if key == "friends[]" {
				friends = value
			} else if key == "amtsplit[]" {
				amtSplit = value
			} else if key == "des" {
				descrption = value[0]
			} else if key == "tamt" {
				totalAmt = value[0]
			} else if key == "billdt" {
				billdt = value[0]
			}
		}
		if len(friends) != len(amtSplit) || len(friends) == 0 || descrption == "" || totalAmt == "" || billdt == "" {
			w.Write([]byte("Invalid data provided"))
		} else {
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(friendsBkt)
				if b == nil {
					return nil
				}
				c := b.Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					tmp, err := gobDecodeFrnd(v)
					checkError(err)
					for index, dat := range friends {
						if tmp.NickName == dat {
							stemp := toSentBill{
								id:     tmp.ID,
								name:   tmp.NickName,
								amount: amtSplit[index],
							}
							sentlist = append(sentlist, stemp)
							break
						}
					}
				}
				return nil
			})
			checkError(err)
			fmt.Println(sentlist)
		}
	}
	w.Write([]byte(""))
}
