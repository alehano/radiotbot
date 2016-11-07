package main

import (
	"encoding/gob"
	"log"
	"os"
)

func LoadShows() *Shows {
	if _, err := os.Stat(showsFilePath); os.IsNotExist(err) {
		return NewShows()
	}

	f, err := os.Open(showsFilePath)
	if err != nil {
		log.Println("Open gob file error: " + err.Error())
		return NewShows()
	}
	dec := gob.NewDecoder(f)
	var shows = NewShows()
	err = dec.Decode(shows)
	if err != nil {
		log.Println("Decode gob error: " + err.Error())
		return NewShows()
	}
	return shows
}

func SaveShows(shows *Shows) error {
	f, err := os.Create(showsFilePath)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(shows)
	if err != nil {
		return err
	}
	return nil
}
