package main

import (
	"log"
	"net/http"

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
	data := struct {
		Active       string
		TotalUsers   int64
		DailyUsers   int64
		LandingUsers map[string]int64
		StateUsers   map[string]int64
	}{
		Active: "messages",
	}

	err := Templates.ExecuteTemplate(w, "admin_messages", data)
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
