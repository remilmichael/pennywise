package main

import (
	"bytes"
	"encoding/binary"
	"log"

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
			log.Panic(err)
		}
	}
	isOpen = false
}
func boltInsert(bkt []byte, key string, dat []byte) error {
	var err error
	/*
		err = boltOpen()
		if err != nil {
			return err
		}
	*/
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bkt)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(key), dat)
		return err
	})
	return err
}

func boltSelect(bkt []byte, key string) ([]byte, error) {
	var err error
	/*
		config := &bolt.Options{Timeout: 1 * time.Second, ReadOnly: true}
		//db, err = bolt.Open("db/data.db", 0600, config)
		//isOpen = true
		if err != nil {
			return nil, err
		}
	*/
	var res []byte
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bkt)
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

func boltBudSearch(bkt []byte, id string, name string) (int8, error) {
	var err error
	var ret int8
	ret = 0
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		var tmp *Friend
		for k, v := c.First(); k != nil; k, v = c.Next() {
			tmp, err = gobDecodeFrnd(v)
			if err != nil {
				return err
			}
			if tmp.ID == id {
				ret++
			}
			if tmp.NickName == name {
				ret = ret + 2
			}
		}
		return nil
	})
	return ret, err
}

func boltReturnAll(bkt []byte) ([]string, error) {
	var err error
	var dat []string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			tmp, err := gobDecodeFrnd(v)
			if err != nil {
				return err
			}
			dat = append(dat, tmp.NickName)
		}
		return nil
	})
	return dat, err
}

func returnIDByName(bkt []byte, name string) (string, error) {
	var err error
	var ID string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			tmp, err := gobDecodeFrnd(v)
			if err != nil {
				return err
			}
			if tmp.NickName == name {
				ID = tmp.ID
				break
			}
		}
		return nil
	})
	return ID, err
}

func queueInsert(bkt []byte, dat []byte) error {
	var err error
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bkt)
		if err != nil {
			return err
		}
		tmp, err := bucket.NextSequence()
		if err != nil {
			checkError(err)
		}
		key := make([]byte, 8)
		binary.LittleEndian.PutUint64(key, uint64(tmp))
		err = bucket.Put(key, dat)
		return nil
	})
	return err
}

func queueCompare(bkt []byte, dat []byte) (bool, error) {
	var err error
	var found bool
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bkt)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if bytes.Equal(v, dat) {
				found = true
				break
			}
		}
		return nil
	})
	return found, err
}
