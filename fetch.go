package main

import (
	"strings"

	"strconv"

	"errors"

	"time"

	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/alehano/batch"
)

const (
	radioTArchiveURL  = radioTURL + "/archives/"
	radioTTitlePrefix = "Радио-Т "
)

// GetShows gets shows data concurrently. Starts from "fromID" (included)
// Returns errors to "errHandler" function.
func GetShows(fromID int, errHandler func(error)) *Shows {
	const workers = 10

	shows := NewShows()
	showsLinks, err := fetchShowsLinks(fromID)
	if err != nil {
		errHandler(err)
		return shows
	}

	batch := batch.New(workers, errHandler)
	batch.Start()

	for _, showURL := range showsLinks {
		fn := func(shows *Shows, url string) func() error {
			return func() error {
				show, err := fetchShow(url)
				if err != nil {
					return err
				} else {
					shows.Add(show)
				}
				return nil
			}
		}
		batch.Add(fn(shows, showURL))
	}

	batch.Close()
	return shows
}

func fetchShowsLinks(fromID int) ([]string, error) {
	links := []string{}

	doc, err := goquery.NewDocument(radioTArchiveURL)
	if err != nil {
		return links, err
	}

	doc.Find("article h1 a").Each(func(i int, s *goquery.Selection) {
		if url, ok := s.Attr("href"); ok {
			text := s.Text()
			if strings.HasPrefix(text, radioTTitlePrefix) {
				numS := strings.TrimPrefix(text, radioTTitlePrefix)
				numS = strings.TrimSpace(numS)
				if num, err := strconv.Atoi(numS); err == nil {
					if fromID > 0 {
						if num >= fromID {
							links = append(links, radioTURL+url)
						}
					} else {
						links = append(links, radioTURL+url)
					}
				}
			}
		}
	})
	return links, nil
}

func fetchShow(url string) (Show, error) {
	show := Show{}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return show, err
	}

	idS := doc.Find("h1.entry-title").Text()
	idS = strings.TrimPrefix(idS, radioTTitlePrefix)
	idS = strings.TrimSpace(idS)
	if id, err := strconv.Atoi(idS); err == nil {
		show.ID = id
	} else {
		return show, errors.New("Bad ID for: " + url)
	}

	show.URL = url

	allErr := []error{}
	doc.Find(".entry-content ul li").Each(func(i int, s *goquery.Selection) {

		topic := s.Text()
		show.TopicsText = append(show.TopicsText, topic)

		// Parse markdown
		s.Each(func(i int, s *goquery.Selection) {
			s.Find("a").Each(func(i int, s *goquery.Selection) {
				if link, ok := s.Attr("href"); ok {
					linkTxt := s.Text()
					mdLink := fmt.Sprintf("[%s](%s)", linkTxt, link)
					topic = strings.Replace(topic, linkTxt, mdLink, 1)
				}
			})
		})
		show.TopicsMarkdown = append(show.TopicsMarkdown, topic)
	})

	if dateS, ok := doc.Find(".meta time").Attr("datetime"); ok {
		if date, err := time.Parse(time.RFC3339, dateS); err == nil {
			show.Date = date
		}
	}

	if image, ok := doc.Find(".entry-content p img").Attr("href"); ok {
		show.ImageURL = image
	}

	doc.Find(".entry-content p a").Each(func(i int, s *goquery.Selection) {
		if link, ok := s.Attr("href"); ok {
			switch s.Text() {
			case "аудио":
				show.AudioURL = link
			case "radio-t.torrent":
				show.TorrentURL = link
			case "лог чата":
				show.ChatLogURL = link
			}
		}
	})

	if len(allErr) > 0 {
		return show, allErr[0]
	}
	return show, nil
}
