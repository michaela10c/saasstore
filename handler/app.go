package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appstore/model"
	"appstore/service"
)

// Parse from body of request to get a json object. Convert to App struct in model.go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    
    fmt.Println("Received one upload request")

		// convert request to model
    decoder := json.NewDecoder(r.Body)
    var app model.App
    if err := decoder.Decode(&app); err != nil {
        panic(err)
    }

		// TODO: backend handle request
		
		// return 
    fmt.Fprintf(w, "Upload request received: %s\n", app.Description)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one search request")
	w.Header().Set("Content-Type", "application/json")
	title := r.URL.Query().Get("title")
	description := r.URL.Query().Get("description")


	var apps []model.App
	var err error
	apps, err = service.SearchApps(title, description)
	if err != nil {
			http.Error(w, "Failed to read Apps from backend", http.StatusInternalServerError)
			return
	}


	js, err := json.Marshal(apps)
	if err != nil {
			http.Error(w, "Failed to parse Apps into JSON format", http.StatusInternalServerError)
			return
	}
	w.Write(js)
}
