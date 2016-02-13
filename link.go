package main

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/google-api-go-client/googleapi/transport"
	urlshortener "google.golang.org/api/urlshortener/v1"
)

type Link struct {
	Hash      string
	UserID    int64
	Action    string
	Payload   map[string]string
	Clicks    int64
	CreatedOn string
	ExpiresIn int64
}

func (this *Link) Save() error {
	db := NewMySQL()

	var err error

	if this.Hash == "" {
		this.Hash = makeHash()

		_, err = db.Insert(
			"INSERT INTO link SET hash=?, user_id=?, action=?, payload=?, expires_in=?",
			this.Hash,
			this.UserID,
			this.Action,
			Stringify(this.Payload),
			this.ExpiresIn,
		)
	} else {
		_, err = db.Update(
			"UPDATE link SET user_id=?, action=?, payload=?, expires_in=? WHERE hash=?",
			this.UserID,
			this.Action,
			Stringify(this.Payload),
			this.ExpiresIn,
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

	if this.Hash == "" {
		return errors.New("Record is missing the hash and can not be loaded.")
	}

	result, err := db.Select("SELECT hash, user_id, action, payload, expires_in, created_on, clicks FROM link WHERE hash=? LIMIT 1", this.Hash)
	if err != nil {
		return err
	}

	for result.Next() {
		var payloadStr string
		var expires sql.NullInt64

		err = result.Scan(&this.Hash, &this.UserID, &this.Action, &payloadStr, &expires, &this.CreatedOn, &this.Clicks)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		this.Payload = Mapify(payloadStr)

		if expires.Valid {
			this.ExpiresIn = expires.Int64
		}
	}

	if this.CreatedOn == "" {
		return errors.New("Hash not found.")
	}

	return nil
}

func (this *Link) Click() error {
	db := NewMySQL()

	_, err := db.Update(
		"UPDATE link SET clicks=clicks+1 WHERE hash=?",
		this.Hash,
	)

	return err
}

func (this *Link) Expire() error {
	db := NewMySQL()

	_, err := db.Update(
		"UPDATE link SET expires_in=1 WHERE hash=?",
		this.Hash,
	)

	return err
}

func (this *Link) IsExpired() (bool, error) {
	if this.CreatedOn == "" {
		this.Load()
	}

	if this.ExpiresIn != 0 {
		loc, _ := time.LoadLocation("Local")
		createdOn, _ := time.ParseInLocation("2006-01-02 15:04:05", this.CreatedOn, loc)

		if time.Now().Local().Unix() > createdOn.Unix()+this.ExpiresIn {
			return true, nil
		}
	}

	return false, nil
}

func (this *Link) Shorten() (string, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: os.Getenv("GOOGLE_API_KEY")},
	}

	svc, err := urlshortener.New(client)
	if err != nil {
		log.Fatalf("Unable to create UrlShortener service: %v", err)
	}

	url, err := svc.Url.Insert(&urlshortener.Url{
		Kind:    "urlshortener#url",
		LongUrl: "http://iwillvote.us/?code=" + this.Hash,
	}).Do()
	if err != nil {
		log.Fatalf("URL Insert: %v", err)
	}

	return url.Id, nil
}

func makeHash() string {
	data := []byte(time.Now().String())
	return fmt.Sprintf("%x", sha1.Sum(data))
}
