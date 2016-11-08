package main

import (
	"fmt"

	"github.com/blevesearch/bleve"
)

func NewSearchIndex() (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	return bleve.NewMemOnly(mapping)
}

func AddToIndex(index bleve.Index, show Show) error {
	for i, topic := range show.TopicsText {
		err := index.Index(fmt.Sprintf("%d:%d", show.ID, i), topic)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReindexAll(index bleve.Index, shows *Shows) error {
	for _, show := range shows.GetItems() {
		err := AddToIndex(index, show)
		if err != nil {
			return err
		}
	}
	return nil
}

func Search(index bleve.Index, q string) error {
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		return err
	}

	fmt.Printf("SEARCH: %s - ", q)
	fmt.Printf("%+v\n", searchResult)

	return nil
}
