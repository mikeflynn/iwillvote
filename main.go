package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var Port = flag.String("port", "8080", "Port for web server to run.")
var WebRoot = flag.String("root", "./webroot/", "The web file root directory.")

func main() {
	flag.Parse()

	// Start message service...
	go sendService()

	// Start email queue handler...
	go EmailSendQueueHandler()

	// Start web server...
	r := mux.NewRouter()

	// Index
	r.HandleFunc("/", pageHandler)

	// Static files
	fs := http.FileServer(http.Dir(*WebRoot + "static"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))

	// API Endpoints
	r.HandleFunc("/api/user/add/", addUserHandler).Methods("POST")

	log.Println("Web server running on " + *Port)

	http.ListenAndServe(":"+*Port, r)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadFile(*WebRoot + "/index.html")
	if err != nil {
		log.Println(err.Error())
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(body)
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var jsonBytes []byte

	r.ParseForm()

	user := &User{
		Network: r.FormValue("network"),
		UUID:    r.FormValue("uuid"),
		Name:    r.FormValue("name"),
		State:   r.FormValue("state"),
	}

	var err error

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
