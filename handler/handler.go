package handler

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	indexer "github.com/qluto/go-full-text-search/indexer"
	searcher "github.com/qluto/go-full-text-search/searcher"
	"io"
	"log"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}

func DumpHandler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	log.Println("dump request")
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		dump := indexer.DumpCursor(tx, c, 0)
		io.WriteString(w, dump)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "GET Method only")
		return
	}
	log.Println("search request")
	io.WriteString(w, searcher.Search(r.FormValue("q"), db))
}

func DeleteHandler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "DELETE Method only")
		return
	}
	log.Println("delete request")
	r.ParseForm()
	params := mux.Vars(r)
	indexer.DeleteDoc(params["id"], db)
}

func IndexingHandler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "PUT Method only")
		return
	}
	log.Println("put request")
	r.ParseForm()
	params := mux.Vars(r)
	indexer.AddDoc(params["id"], r.FormValue("body"), db)
}
