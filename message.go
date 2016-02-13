package main

import (
	"errors"
	"fmt"
)

type Message struct {
	ID        int64        `json:"id"`
	To        []*MessageTo `json:"to"`
	Slug      string       `json:"slug"`
	Message   string       `json:"message"`
	Outgoing  int          `json:"outgoing"`
	CreatedOn string       `json:"created_on"`
	SendOn    string       `json:"send_on"`
	Sent      int          `json:"sent"`
}

func GetMessagesToSend() ([]*Message, error) {
	db := NewMySQL()

	result, err := db.Select(`SELECT m.id, network, uuid, slug, message, outgoing, m.created_on, send_on, sent
		FROM user_message AS um
		LEFT JOIN message AS m ON (m.id = um.message_id)
		WHERE sent = 0 AND send_on < now()`)
	if err != nil {
		return []*Message{}, err
	}

	rows := []*Message{}

	for result.Next() {
		msg := &Message{}
		msgTo := &MessageTo{}

		result.Scan(&msg.ID, &msgTo.Network, &msgTo.UUID, &msg.Slug, &msg.Message, &msg.Outgoing, &msg.CreatedOn, &msg.SendOn, &msg.Sent)
		msg.To = []*MessageTo{msgTo}
		rows = append(rows, msg)
	}

	return rows, nil
}

func (this *Message) Save() error {
	if this.Message == "" || this.Slug == "" {
		return errors.New("Missing required message and slug fields.")
	}

	db := NewMySQL()

	var err error

	// Message Table Record
	if this.ID == 0 {
		newID, err := db.Insert(
			"INSERT INTO message SET slug=?, message=?, outgoing=?",
			this.Slug,
			this.Message,
			this.Outgoing,
		)

		if err == nil {
			this.ID = newID
		}
	} else {
		_, err = db.Update(
			"UPDATE message SET slug=?, message=?, outgoing=? WHERE id=?",
			this.Slug,
			this.Message,
			this.Outgoing,
			this.ID,
		)
	}

	if err != nil {
		return err
	}

	// UserMessage Table Record
	for _, um := range this.To {
		um.MessageID = this.ID

		if um.SendOn == "" {
			um.SendOn = this.SendOn
		}

		err = um.Save()
	}

	if err != nil {
		return err
	}

	return nil
}

func (this *Message) AddTo(uuid string, network string) {
	this.To = append(this.To, &MessageTo{
		UUID:      uuid,
		Network:   network,
		MessageID: this.ID,
	})
}

func (this *Message) Load() error {
	db := NewMySQL()

	params := []interface{}{}
	where := ""

	if this.ID > 0 {
		where = "id=?"
		params = append(params, this.ID)
	} else if this.Slug != "" {
		where = "slug=?"
		params = append(params, this.Slug)
	} else {
		return errors.New("Message missing required fields for load: id")
	}

	result, err := db.Select("SELECT id, message, slug, outgoing, created_on FROM message WHERE "+where+" LIMIT 1", params...)
	if err != nil {
		return err
	}

	for result.Next() {
		result.Scan(&this.ID, &this.Message, &this.Slug, &this.Outgoing, &this.CreatedOn)
	}

	return nil
}

func (this *Message) LoadTo(per int, page int) error {
	db := NewMySQL()

	offset := (per * page) - per

	result, err := db.Select("SELECT id, message_id, network, uuid, send_on, sent FROM user_message WHERE message_id=? LIMIT ?,?", this.ID, offset, page)
	if err != nil {
		return err
	}

	for result.Next() {
		mt := &MessageTo{}
		result.Scan(&mt.ID, &mt.MessageID, &mt.Network, &mt.UUID, &mt.SendOn, &mt.Sent)

		this.To = append(this.To, mt)
	}

	return nil
}

func (this *Message) Send() error {
	if this.ID == 0 {
		this.Save()
	}

	var errArr []error = nil

	for _, mt := range this.To {
		mt.MessageID = this.ID
		if err := mt.Send(this); err != nil {
			errArr = append(errArr, err)
		}
	}

	if errArr != nil {
		errMsg := ""
		for _, x := range errArr {
			errMsg += x.Error()
		}

		return errors.New(errMsg)
	}

	return nil
}

type MessageTo struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"message_id"`
	Network   string `json:"network"`
	UUID      string `json:"uuid"`
	SendOn    string `json:"send_on"`
	Sent      int    `json:"sent"`
}

func (this *MessageTo) Save() error {
	db := NewMySQL()

	var err error

	// Message Table Record
	if this.ID == 0 {
		newID, err := db.Insert(
			"INSERT INTO user_message SET message_id=?, network=?, uuid=?, send_on=?, sent=?",
			this.MessageID,
			this.Network,
			this.UUID,
			SQLNullIfEmpty(this.SendOn),
			this.Sent,
		)

		if err == nil {
			this.ID = newID
		}
	} else {
		_, err = db.Update(
			"UPDATE user_message SET message_id=?, network=?, uuid=?, send_on=?, sent=? WHERE id=?",
			this.MessageID,
			this.Network,
			this.UUID,
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

func (this *MessageTo) Load() error {
	db := NewMySQL()

	params := []interface{}{}
	where := ""

	if this.ID > 0 {
		where = "id=?"
		params = append(params, this.ID)
	} else {
		return errors.New("Message missing required fields for load: id")
	}

	result, err := db.Select("SELECT id, message_id, network, uuid, send_on, sent FROM message WHERE "+where+" LIMIT 1", params...)
	if err != nil {
		return err
	}

	for result.Next() {
		result.Scan(&this.ID, &this.MessageID, &this.Network, &this.UUID, &this.SendOn, &this.Sent)
	}

	return nil
}

func (this *MessageTo) Send(msg *Message) error {
	var err error

	if err = this.Email(msg); err == nil {
		this.Sent = 1

		if err = this.Save(); err == nil {
			return nil
		}
	}

	return err
}

func (this *MessageTo) Email(msg *Message) error {
	email := &Email{
		From:    "sms@iwillvote.us",
		To:      fmt.Sprintf(NetworkToDomain(this.Network), this.UUID),
		Subject: "",
		Body:    msg.Message,
	}

	if err := email.Send(); err != nil {
		return err
	}

	return nil
}

var messageDomains map[string]string = map[string]string{
	"att":        "%s@txt.att.net",
	"metropcs":   "%s@mymetropcs.com",
	"sprint":     "%s@messaging.sprintpcs.com",
	"tmobile":    "%s@tmomail.net",
	"tracfone":   "%s@txt.att.net",
	"uscellular": "%s@email.uscc.net",
	"verizon":    "%s@vtext.com",
	"virgin":     "%s@vmobl.com",
	"gmail":      "%s@gmail.com",
}

func NetworkToDomain(network string) string {
	if v, ok := messageDomains[network]; ok {
		return v
	}

	return ""
}

func DomainToNetwork(domain string) string {
	for network, match := range messageDomains {
		if match == "%s@"+domain {
			return network
		}
	}

	return ""
}
