package main

import (
	"time"

	"github.com/blevesearch/bleve"

	"log"

	"net/http"

	"gopkg.in/robfig/cron.v2"
)

const (
	radioTURL     = "https://radio-t.com"
	showsFilePath = "./data/shows.gob"
)

var searchIndex bleve.Index

func main() {
	shows := LoadShows()

	newSearchIndex, err := NewSearchIndex()
	if err != nil {
		log.Fatal("Search index create error:", err)
	}
	searchIndex = newSearchIndex
	err = ReindexAll(searchIndex, shows)
	if err != nil {
		log.Fatal("Search reindex error:", err)
	}

	// Update if last show was more than 7 days ago
	if shows.Last().Date.Add(7 * 24 * time.Hour).Before(time.Now()) {
		Update(shows, searchIndex)
	}

	// Update every Monday at 6:00
	c := cron.New()
	_, err = c.AddFunc("0 6 * * * 1", func() {
		Update(shows, searchIndex)
	})
	if err != nil {
		log.Fatal("Add to cron error:", err)
	}

	// TODO: panic recover!!!

	// Run server
	http.HandleFunc("/test/", SearchHandler)
	http.ListenAndServe(":8082", nil)
}
