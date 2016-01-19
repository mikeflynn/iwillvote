package main

import (
	"errors"
	"fmt"
	"os"
)

type Message struct {
	ID        int64  `json:"id"`
	Network   string `json:"network"`
	UUID      string `json:"uuid"`
	Message   string `json:"message"`
	Outgoing  int    `json:"outgoing"`
	CreatedOn string `json:"created_on"`
	SendOn    string `json:"send_on"`
	Sent      int    `json:"sent"`
}

func GetMessagesToSend() ([]*Message, error) {
	db := NewMySQL()

	result, err := db.Select("SELECT id, network, uuid, message, outgoing, created_on, send_on, sent FROM message WHERE sent = 0 AND send_on < now()")
	if err != nil {
		return []*Message{}, err
	}

	rows := []*Message{}

	for result.Next() {
		msg := &Message{}
		result.Scan(&msg.ID, &msg.Network, &msg.UUID, &msg.Message, &msg.Outgoing, &msg.CreatedOn, &msg.SendOn, &msg.Sent)
		rows = append(rows, msg)
	}

	return rows, nil
}

func (this *Message) Save() error {
	db := NewMySQL()

	var err error

	if this.ID == 0 {
		newID, err := db.Insert(
			"INSERT INTO message SET network=?, uuid=?, message=?, outgoing=?, send_on=?, sent=?",
			this.Network,
			this.UUID,
			this.Message,
			this.Outgoing,
			SQLNullIfEmpty(this.SendOn),
			this.Sent,
		)

		if err == nil {
			this.ID = newID
		}
	} else {
		_, err = db.Update(
			"UPDATE message SET network=?, uuid=?, message=?, outgoing=?, send_on=?, sent=? WHERE id=?",
			this.Network,
			this.UUID,
			this.Message,
			this.Outgoing,
			SQLNullIfEmpty(this.SendOn),
			this.Sent,
			this.ID,
		)
	}

	if err != nil {
		return err
	}

	return nil
}

func (this *Message) Load() error {
	db := NewMySQL()

	params := []interface{}{}
	where := ""

	if this.ID > 0 {
		where = "id=?"
		params = append(params, this.ID)
	} else if this.Network != "" && this.UUID != "" {
		where = "network=? AND uuid=?"
		params = append(params, this.Network, this.UUID)
	} else {
		return errors.New("Message missing required fields for load: id or network and uuid")
	}

	result, err := db.Select("SELECT id, network, uuid, message, outgoing, created_on, send_on, sent FROM message WHERE "+where+" LIMIT 1", params...)
	if err != nil {
		return err
	}

	for result.Next() {
		result.Scan(&this.ID, &this.Network, &this.UUID, &this.Message, &this.Outgoing, &this.CreatedOn, &this.SendOn, &this.Sent)
	}

	return nil
}

func (this *Message) Send() error {
	var err error

	if err = this.Email(); err == nil {
		this.Sent = 1

		if err = this.Save(); err == nil {
			return nil
		}
	}

	return err
}

func (this *Message) Email() error {
	domains := map[string]string{
		"att":        "%s@txt.att.net",
		"metropcs":   "%s@mymetropcs.com",
		"sprint":     "%s@messaging.sprintpcs.com",
		"tmobile":    "%s@tmomail.net",
		"tracfone":   "%s@txt.att.net",
		"uscellular": "%s@email.uscc.net",
		"verizon":    "%s@vtext.com",
		"virgin":     "%s@vmobl.com",
	}

	if os.Getenv("GMAIL_ADDRESS") == "" || os.Getenv("GMAIL_PASSWORD") == "" {
		return errors.New("GMail not configured.")
	}

	email := &Email{
		From:     os.Getenv("GMAIL_ADDRESS"),
		Password: os.Getenv("GMAIL_PASSWORD"),
		To:       fmt.Sprintf(domains[this.Network], this.UUID),
		Subject:  "",
		Body:     this.Message,
	}

	if err := email.Send(); err != nil {
		return err
	}

	return nil
}
