package main

import (
	"errors"
	"log"
	"regexp"
)

func GetUserCount() (int64, error) {
	db := NewMySQL()

	var count int64

	result, err := db.Select("SELECT count(*) FROM user")
	if err != nil {
		return count, err
	}

	for result.Next() {
		err := result.Scan(&count)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func GetUsersCountByCandidate() (map[string]int64, error) {
	db := NewMySQL()

	var count map[string]int64

	result, err := db.Select("SELECT IF(landing_page = '', 'front', landing_page) AS landing_page, count(*) AS total FROM user GROUP BY landing_page")
	if err != nil {
		return count, err
	}

	for result.Next() {
		var p string
		var c int64
		err := result.Scan(&p, &c)
		if err != nil {
			return count, err
		}

		count[p] = c
	}

	return count, nil
}

func ListUsers(landing string, state string, sort string, limit int) ([]*User, error) {
	db := NewMySQL()

	var userList []*User

	where := ""
	whereVars := []string{}

	if landing != "" {
		where += "landing_page = ?"
		whereVars = append(whereVars, landing)
	}

	if state != "" {
		where += "state = ?"
		whereVars = append(whereVars, state)
	}

	result, err := db.Select(`SELECT
		id, network, uuid, name, state, zipcode, created_on, deleted, landing_page, message_window, news, reminders
		FROM user WHERE ? ORDER BY ? DESC LIMIT ?`,
		where, sort, limit)
	if err != nil {
		return userList, err
	}

	for result.Next() {
		u := &User{}
		err := result.Scan(&u.ID, &u.Network, &u.UUID, &u.Name, &u.State, &u.Zipcode, &u.CreatedOn, &u.Deleted, &u.LandingPage, &u.MessageWindow, &u.News, &u.Reminders)
		if err != nil {
			return userList, err
		}

		userList = append(userList, u)
	}

	return userList, nil
}

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

func (this *User) IsComplete() error {
	if this.Network != "" && this.UUID != "" {
		_, err := regexp.MatchString("^\\d{10}$", this.UUID)
		if err == nil {
			return nil
		} else {
			return errors.New("Invalid phone number UUID.")
		}
	} else {
		return errors.New("Missing one or more required fields.")
	}
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

	if this.CreatedOn == "" || this.Deleted == 1 {
		return errors.New("User not found or deleted.")
	}

	return nil
}

func (this *User) Unsubscribe() error {
	this.Deleted = 1
	err := this.Save()
	return err
}
