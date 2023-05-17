package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"appstore/model"
	"appstore/service"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

// Parse from body of request to get a json object. Convert to App struct in model.go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one upload request")

	// Parse from body of request to get a json object.
	fmt.Println("Received one upload request")

	token := r.Context().Value("user")
	claims := token.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"]

	app := model.App{
			Id:          uuid.New(),
			User:        username.(string),
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
	}
	
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	fmt.Printf("%v,%T", price, price)
	if err != nil {
			fmt.Println(err)
	}
	app.Price = int(price * 100.0)

	
	file, _, err := r.FormFile("media_file")
	if err != nil {
			http.Error(w, "Media file is not available", http.StatusBadRequest)
			fmt.Printf("Media file is not available %v\n", err)
			return
	}

	// backend handle request + error handling
	err = service.SaveApp(&app, file)
	if err != nil {
			http.Error(w, "Failed to save app to backend", http.StatusInternalServerError)
			fmt.Printf("Failed to save app to backend %v\n", err)
			return
	}
	
	fmt.Fprintf(w, "App is saved successfully: %s\n", app.Description)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one search request")

	// set return type to json 
	w.Header().Set("Content-Type", "application/json")

	// get params from query
	title := r.URL.Query().Get("title")
	description := r.URL.Query().Get("description")

	// 
	var apps []model.App
	var err error
	apps, err = service.SearchApps(title, description)
	if err != nil {
			http.Error(w, "Failed to read Apps from backend", http.StatusInternalServerError)
			return
	}

	// construct return
	js, err := json.Marshal(apps)
	if err != nil {
			http.Error(w, "Failed to parse Apps into JSON format", http.StatusInternalServerError)
			return
	}
	w.Write(js)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one checkout request")

	// set return type to plain text (checkout session url) 
	w.Header().Set("Content-Type", "text/plain")

	// 
	appID := r.FormValue("appID")
	url, err := service.CheckoutApp(r.Header.Get("Origin"), appID)
	if err != nil {
			fmt.Println("Checkout failed.")
			w.Write([]byte(err.Error()))
			return
	}

	// construct return
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))

	// notify start of checkout process
	fmt.Println("Checkout process started!")
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one request for delete")

	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"].(string)
	id := mux.Vars(r)["id"]

	if err := service.DeleteApp(id, username); err != nil {
		http.Error(w, "Failed to delete app from backend", http.StatusInternalServerError)
		fmt.Printf("Failed to delete app from backend %v\n", err)
		return
	}

	fmt.Println("App is deleted successfully")
}


