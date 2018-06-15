package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/metalblueberry/dashboard/app"
)

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
		Handle("/login.html", http.FileServer(http.Dir("page/open/"))).
		Methods("GET")

	router.
		PathPrefix("/").
		Handler(app.WithAuth(http.FileServer(http.Dir("page/")))).
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
