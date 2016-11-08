package search

import (
	"fmt"

	"github.com/alehano/radiotbot/shows"
	"github.com/blevesearch/bleve"
)

func NewIndex() (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	return bleve.NewMemOnly(mapping)
}

func AddToIndex(index bleve.Index, show shows.Show) error {
	for i, topic := range show.TopicsText {
		err := index.Index(fmt.Sprintf("%d:%d", show.ID, i), topic)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReindexAll(index bleve.Index, shows *shows.Shows) error {
	for _, show := range shows.GetItems() {
		err := AddToIndex(index, show)
		if err != nil {
			return err
		}
	}
	return nil
}

func Query(index bleve.Index, q string) (string, error) {
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		return "", err
	}

	fmt.Printf("SEARCH: %s - ", q)
	fmt.Printf("%+v\n", searchResult)

	return "", nil
}
