package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
)

// CLI Params
var Port = flag.String("port", "8080", "Port for web server to run.")
var WebRoot = flag.String("root", "./webroot/", "The web file root directory.")

// Templates
var Templates *template.Template

// Candidates
var Candidates map[string]bool = map[string]bool{
	"clinton":  true,
	"sanders":  true,
	"cruz":     true,
	"carson":   true,
	"fiorina":  true,
	"christie": true,
	"trump":    true,
	"rubio":    true,
	"bush":     true,
	"kodos":    true,
}

func main() {
	flag.Parse()

	// Load Templates
	Templates = template.Must(template.ParseGlob(*WebRoot + "/templates/*"))

	log.Printf("Web root directory set to: %s", *WebRoot)

	// Start message services...
	go sendService()
	go receiveService()

	// Start web server...
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(*WebRoot)))

	// API Endpoints
	r.HandleFunc("/api/user/add/", addUserHandler).Methods("POST")
	r.HandleFunc("/api/user/remove/", removeUserHandler).Methods("POST")

	// Admin Endpoints
	ar := mux.NewRouter().PathPrefix("/admin").Subrouter()
	ar.HandleFunc("/", adminHandler)
	ar.HandleFunc("/api", adminHandler)
	r.PathPrefix("/admin").Handler(httpauth.SimpleBasicAuth("user", "pass")(ar))

	// Pages
	r.HandleFunc("/unsubscribe", unsubHandler).Methods("POST", "GET")
	r.HandleFunc("/code/{code:[0-9a-f]{5,40}}", codeHandler)
	r.HandleFunc("/{page:[a-z]*}", pageHandler)
	log.Println("Web server running on " + *Port)

	http.ListenAndServe(":"+*Port, r)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	candidate := ""
	page := "index"
	if v, ok := params["page"]; ok {
		if v != "" {
			if Candidates[v] {
				candidate = v
			} else {
				page = v
			}
		}
	}

	data := struct {
		Title         string
		Active        string
		Candidate     string
		CandidateList map[string]bool
	}{
		Title:         "i Will Vote",
		Active:        page,
		Candidate:     candidate,
		CandidateList: Candidates,
	}

	err := Templates.ExecuteTemplate(w, page, data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ADMIN!\n"))
}

func codeHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var message, errorText string

	params := mux.Vars(r)

	link := &Link{
		Hash: params["code"],
	}

	err = link.Load()
	expired, _ := link.IsExpired()
	if err != nil || expired {
		log.Println(fmt.Sprintf("Code not found: %s", params["code"]))
		http.NotFound(w, r)
		return
	}

	link.Click()

	switch link.Action {
	case "unsubscribe":
		user := &User{ID: link.UserID}
		if err = user.Load(); err == nil {
			if err = user.Unsubscribe(); err == nil {
				message = "You have successfully been unsubscribed! Please remember to vote a different way."

				link.Expire()
			}
		}
	}

	data := struct {
		Title         string
		Active        string
		Candidate     string
		CandidateList map[string]bool
		Message       string
		Error         string
	}{
		Title:         "i Will Vote",
		Active:        "",
		Candidate:     "",
		CandidateList: Candidates,
		Message:       message,
		Error:         errorText,
	}

	err = Templates.ExecuteTemplate(w, "code", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}

func unsubHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	message := ""
	errorText := ""

	err = r.ParseForm()
	if err == nil && r.FormValue("uuid") != "" {
		user := &User{
			UUID:    r.FormValue("uuid"),
			Network: r.FormValue("network"),
		}

		if err = user.IsComplete(); err == nil {
			if err = user.Load(); err == nil {
				link := &Link{
					Action:    "unsubscribe",
					UserID:    user.ID,
					ExpiresIn: 3600,
				}

				if err = link.Save(); err == nil {
					msg := &Message{
						Network:  user.Network,
						UUID:     user.UUID,
						Message:  fmt.Sprintf("We have received a request to unsubscribe you from iWillVote.us. Tap on this link to complete the process http://iwillvote.us/code/%s, or do nothing if you don't want to unsubscribe.", link.Hash),
						Outgoing: 1,
					}

					if err = msg.Send(); err == nil {
						message = "If this matches a user in our system we will send a verification link to complete your request."
					}
				}
			}
		}
	}

	if err != nil {
		log.Println(err.Error())
		message = "If this matches a user in our system we will send a verification link to complete your request."
		//errorText = "We couldn't complete your unsubscribe action at this time. Please try again in a moment."
	}

	data := struct {
		Title         string
		Active        string
		Candidate     string
		CandidateList map[string]bool
		Message       string
		Error         string
	}{
		Title:         "i Will Vote",
		Active:        "unsubscribe",
		Candidate:     "",
		CandidateList: Candidates,
		Message:       message,
		Error:         errorText,
	}

	err = Templates.ExecuteTemplate(w, "unsubscribe", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}

func removeUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var jsonBytes []byte

	r.ParseForm()

	user := &User{
		Network: r.FormValue("network"),
		UUID:    r.FormValue("uuid"),
	}

	if err = user.IsComplete(); err != nil {
		jsonBytes, _ = json.Marshal(webError{Error: err.Error()})
	}

	err = user.Load()
	if user.ID != 0 {
		user.Deleted = 1
		if err = user.Save(); err != nil {
			log.Println(err.Error())
			jsonBytes, _ = json.Marshal(webError{Error: "Unable to update user."})
		} else {
			jsonBytes, _ = json.Marshal(webError{Error: "User removed."})
		}
	} else {
		jsonBytes, _ = json.Marshal(webError{Error: "Unable to locate user."})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var jsonBytes []byte

	r.ParseForm()

	user := &User{
		Network:       r.FormValue("network"),
		UUID:          r.FormValue("uuid"),
		Name:          r.FormValue("name"),
		State:         r.FormValue("state"),
		MessageWindow: r.FormValue("window"),
		LandingPage:   r.FormValue("landing_page"),
	}

	if err = user.IsComplete(); err != nil {
		jsonBytes, _ = json.Marshal(webError{Error: err.Error()})
	}

	err = user.Load()
	if user.ID == 0 {
		if err = user.Save(); err == nil {
			message := &Message{
				Network:  user.Network,
				UUID:     user.UUID,
				Message:  "Thanks for signing up! We'll remind you when to vote. Head to iwillvote.us for any questions.",
				Outgoing: 1,
			}

			if err = message.Save(); err == nil {
				if err = message.Send(); err == nil {
					jsonBytes, _ = json.Marshal(webUserResponse{Data: []*User{user}, Status: "User created and welcome message sent."})
				}
			}
		}

		if err != nil {
			log.Println(err.Error())
			jsonBytes, _ = json.Marshal(webError{Error: "Couldn't create user or send welcome message."})
		}
	} else {
		jsonBytes, _ = json.Marshal(webError{Error: "User already exists."})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

type webError struct {
	Error string `json:"error"`
}

type webUserResponse struct {
	Data   []*User
	Status string
}

func sendService() {
	// Start email queue handler...
	go EmailSendQueueHandler()

	// Load it up!
	for {
		log.Println("Getting scheduled messages.")

		rows, err := GetMessagesToSend()
		if err != nil {
			log.Println(err.Error())
		}

		log.Printf("Found %d scheduled messages.\n", len(rows))

		for _, msg := range rows {
			msg.Send()
		}

		time.Sleep(1000 * time.Millisecond * 60 * 10) // 10 minutes
	}
}

func receiveService() {
	for {
		var count int
		var err error

		log.Println("Checking for received messages.")

		if count, err = ProcessS3Emails("iwillvote-sms", 10); err != nil {
			log.Println(err.Error())
		}

		log.Printf("Found %d messages.\n", count)

		time.Sleep(1000 * time.Millisecond * 60 * 10) // 5 minutes
	}
}
