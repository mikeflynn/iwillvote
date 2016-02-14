package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	page := "index"
	if v, ok := params["page"]; ok {
		if v != "" {
			page = v
		}
	}

	data := struct {
		Active string
	}{
		Active: page,
	}

	err := Templates.ExecuteTemplate(w, "admin_index", data)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}
}
