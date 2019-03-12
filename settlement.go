package main

import (
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
