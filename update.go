package main

import (
	"log"

	"sort"

	"github.com/blevesearch/bleve"
)

func Update(shows *Shows, index bleve.Index) {
	log.Println("Updating...")
	newShows := GetShows(shows.Last().ID, func(err error) {
		log.Println(err)
	})

	for _, show := range newShows.GetItems() {
		shows.Add(show)
		err := AddToIndex(index, show)
		if err != nil {
			log.Println(err)
		}
	}

	sort.Sort(shows)

	err := SaveShows(shows)
	if err != nil {
		log.Println(err)
	}
	log.Println("Updated")
}
