package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"time"
)

type Link struct {
	Hash      string
	UserID    int64
	Action    string
	Payload   map[string]string
	CreatedOn string
}

func (this *Link) Save() error {
	db := NewMySQL()

	var err error

	if this.Hash == "" {
		this.Hash = makeHash()

		_, err = db.Insert(
			"INSERT INTO user SET hash=?, user_id=?, action=?, payload=?",
			this.Hash,
			this.UserID,
			this.Action,
			this.Payload,
		)
	} else {
		_, err = db.Update(
			"UPDATE link SET user_id=?, action=?, payload=? WHERE hash=?",
			this.UserID,
			this.Action,
			this.Payload,
			this.Hash,
		)
	}

	if err != nil {
		return err
	}

	return nil
}

func (this *Link) Load() error {
	db := NewMySQL()

	result, err := db.Select("SELECT hash, user_id, action, payload, created_on FROM link WHERE hash=? LIMIT 1", this.Hash)
	if err != nil {
		return err
	}

	for result.Next() {
		err = result.Scan(&this.Hash, &this.UserID, &this.Action, &this.Payload, &this.CreatedOn)
		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	return nil
}

func makeHash() string {
	data := []byte(time.Now().String())
	return fmt.Sprintf("%x", sha1.Sum(data))
}
