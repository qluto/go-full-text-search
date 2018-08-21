package main

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	handler "github.com/qluto/go-full-text-search/handler"
	"log"
	"net/http"
)

var db *bolt.DB

func init() {
}

func main() {
	db, err := bolt.Open("./index.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", handler.RootHandler).Methods("GET")

	router.HandleFunc("/document/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.IndexingHandler(w, r, db)
	}).Methods("PUT")
	router.HandleFunc("/document/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.DeleteHandler(w, r, db)
	}).Methods("DELETE")
	router.HandleFunc("/document", func(w http.ResponseWriter, r *http.Request) {
		handler.DumpHandler(w, r, db)
	}).Methods("GET")

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		handler.SearchHandler(w, r, db)
	}).Methods("GET")

	http.ListenAndServe(":8080", router)
}
