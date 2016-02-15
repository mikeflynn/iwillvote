package main

import (
	"log"
	"net/http"
	"strings"

	//"github.com/gorilla/mux"
)

func AdminIndexHandler(w http.ResponseWriter, r *http.Request) {
	// All-time users
	totalUsers, _ := GetUserCount("")

	// 24-hour users
	todayUsers, _ := GetUserCount("-24h")

	// Candidate Users
	landingUsers, _ := GetUsersCountByLanding()

	// State Users
	stateUsers, _ := GetUsersCountByState()

	data := struct {
		Active       string
		TotalUsers   int64
		DailyUsers   int64
		LandingUsers map[string]int64
		StateUsers   map[string]int64
	}{
		Active:       "index",
		TotalUsers:   totalUsers,
		DailyUsers:   todayUsers,
		LandingUsers: landingUsers,
		StateUsers:   stateUsers,
	}

	err := Templates.ExecuteTemplate(w, "admin_index", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}

func AdminMessagesHandler(w http.ResponseWriter, r *http.Request) {
	var errorMsg, successMsg string

	err := r.ParseForm()
	if err == nil && r.FormValue("slug") != "" {
		msg := &Message{
			Slug:     strings.ToLower(r.FormValue("slug")),
			Message:  r.FormValue("body"),
			Outgoing: 1,
		}

		if err := msg.Save(); err != nil {
			log.Println(err.Error())
			errorMsg = "Unable to create message."
		} else {
			successMsg = "Message created!"
		}
	}

	messageList, err := GetMessageList()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error.", 500)
		return
	}

	data := struct {
		Active      string
		MessageList []*Message
		Success     string
		Error       string
	}{
		Active:      "messages",
		MessageList: messageList,
		Success:     successMsg,
		Error:       errorMsg,
	}

	err = Templates.ExecuteTemplate(w, "admin_messages", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}

func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Active       string
		TotalUsers   int64
		DailyUsers   int64
		LandingUsers map[string]int64
		StateUsers   map[string]int64
	}{
		Active: "users",
	}

	err := Templates.ExecuteTemplate(w, "admin_users", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}
