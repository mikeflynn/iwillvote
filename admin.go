package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

func AdminUserDetailHandler(w http.ResponseWriter, r *http.Request) {
	//var username string

	vars := mux.Vars(r)
	if _, ok := vars["user"]; !ok {
		http.NotFound(w, r)
		return
	}
}

func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	var landing, state string

	if v := params.Get("landing"); v != "" {
		landing = v
	}

	if v := params.Get("state"); v != "" {
		state = v
	}

	var limit int64 = 20
	var offset int64 = 0
	var page int64 = 1
	var prevLink string = "#"
	var nextLink string = "#"
	if v := params.Get("page"); v != "" {
		page, _ = strconv.ParseInt(v, 10, 64)
		offset = (page * limit) - limit
	}

	if offset != 0 {
		prevQuery := r.URL.Query()
		prevQuery.Set("page", strconv.FormatInt(page-1, 10))
		prevLink = "?" + prevQuery.Encode()
	}

	nextQuery := r.URL.Query()
	nextQuery.Set("page", strconv.FormatInt(page+1, 10))
	nextLink = "?" + nextQuery.Encode()

	userList, err := ListUsers(landing, state, "", limit, offset)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error.", 500)
		return
	}

	data := struct {
		Active      string
		MessageList []*Message
		UserList    []*User
		Success     string
		Error       string
		Prev        string
		Next        string
	}{
		Active:   "users",
		UserList: userList,
		Prev:     prevLink,
		Next:     nextLink,
	}

	err = Templates.ExecuteTemplate(w, "admin_users", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}
