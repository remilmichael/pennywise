package main

import (
	"fmt"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var isOpen bool

func boltOpen() error {
	isOpen = false
	var err error
	if !isOpen {
		//config := &bolt.Options{Timeout: 1 * time.Second}
		db, err = bolt.Open("db/data.db", 0600, nil)
	}
	if err != nil {
		return err
	}
	isOpen = true
	return nil
}

func boltClose() {
	if isOpen {
		err := db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
	isOpen = false
}
func boltInsert(bkt string, key string, dat []byte) error {
	var err error
	/*
		err = boltOpen()
		if err != nil {
			return err
		}
	*/
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bkt))
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(key), dat)
		return err
	})
	return err
}

func boltSelect(bkt string, key string) ([]byte, error) {
	var err error
	//config := &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true}
	//db, err = bolt.Open("db/data.db", 0600, config)
	//isOpen = true
	if err != nil {
		return nil, err
	}
	var res []byte
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bkt))
		if bucket == nil {
			return nil
		}
		key := []byte(key)
		_, _ = key, bucket
		res = bucket.Get(key)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
