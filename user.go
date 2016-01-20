package main

import (
	"errors"
	"log"
)

type User struct {
	ID            int64  `json:"id"`
	Network       string `json:"network"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	State         string `json:"state"`
	Zipcode       int    `json:"zipcode"`
	CreatedOn     string `json:"created_on"`
	Deleted       int    `json:"deleted"`
	LandingPage   string `json:"landing_page"`
	MessageWindow string `json:"message_window"`
	News          int    `json:"news_feed"`
	Reminders     int    `json:"reminders"`
}

//func FilterUsers() ([]*User, error) {
//
//}

func (this *User) IsComplete() bool {
	if this.Network != "" && this.UUID != "" && this.Name != "" && this.State != "" {
		return true
	}

	return false
}

func (this *User) Save() error {
	db := NewMySQL()

	var err error

	if this.ID == 0 {
		newID, err := db.Insert(
			"INSERT INTO user SET network=?, uuid=?, name=?, state=?, zipcode=?, deleted=?, landing_page=?, message_window=?, news=?, reminders=?",
			this.Network,
			this.UUID,
			this.Name,
			this.State,
			this.Zipcode,
			this.Deleted,
			this.LandingPage,
			this.MessageWindow,
			this.News,
			this.Reminders,
		)

		if err == nil {
			this.ID = newID
		}
	} else {
		_, err = db.Update(
			"UPDATE user SET network=?, uuid=?, name=?, state=?, zipcode=?, deleted=?, landing_page=?, message_window=?, news=?, reminders=? WHERE id=?",
			this.Network,
			this.UUID,
			this.Name,
			this.State,
			this.Zipcode,
			this.Deleted,
			this.LandingPage,
			this.MessageWindow,
			this.News,
			this.Reminders,
			this.ID,
		)
	}

	if err != nil {
		return err
	}

	return nil
}

func (this *User) Load() error {
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

	result, err := db.Select("SELECT id, network, uuid, name, state, zipcode, created_on, deleted, landing_page, message_window, news, reminders FROM user WHERE "+where+" LIMIT 1", params...)
	if err != nil {
		return err
	}

	for result.Next() {
		err = result.Scan(&this.ID, &this.Network, &this.UUID, &this.Name, &this.State, &this.Zipcode, &this.CreatedOn, &this.Deleted, &this.LandingPage, &this.MessageWindow, &this.News, &this.Reminders)
		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	return nil
}
