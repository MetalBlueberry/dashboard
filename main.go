package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/metalblueberry/dashboard/app"
	"github.com/metalblueberry/dashboard/reports"
	"github.com/metalblueberry/dashboard/reports/table"
	"github.com/metalblueberry/dashboard/reports/timeline"
)

var Services map[string]reports.Report

func init() {
	ReloadReports()
}
func ReloadReports() {
	file, err := ioutil.ReadFile("reports.json")
	if err != nil {
		log.Print(err)
	}
	if len(file) == 0 {
		return
	}
	var ServicesData map[string]interface{}
	err = json.Unmarshal(file, &ServicesData)
	if err != nil {
		log.Print(err)
	}

	Services = make(map[string]reports.Report)
	for Service, data := range ServicesData {
		datastruct := data.(map[string]interface{})
		switch datastruct["type"] {
		case "timeline":
			Services[Service] = timeline.LoadFromData(datastruct["data"])
		}

	}

	log.Printf("Reloaded reports.json")
}

func main() {
	log.Printf("Start code")

	router := mux.NewRouter()

	api := router.PathPrefix("/api").Subrouter()
	api.
		HandleFunc("/login", app.Login).
		Methods("POST")
	api.
		HandleFunc("/logout", app.Logout).
		Methods("GET")

	router.
		PathPrefix("/login/").
		Handler(app.TemplateServer()).
		Methods("GET")

	router.
		PathPrefix("/static/").
		Handler(http.FileServer(http.Dir("page/"))).
		Methods("GET")
	router.
		PathPrefix("/csv/").
		Handler(table.ServeCSVFile("page/reports/table/before.html", "page/reports/table/after.html")).
		Methods("GET")
	router.
		PathPrefix("/").
		Handler(app.WithAuth(app.TemplateServer())).
		Methods("GET")

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	srv := &http.Server{
		Handler: loggedRouter,
		Addr:    ":4430",

		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Starting server")
	log.Fatal(srv.ListenAndServe())
	//log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
}
