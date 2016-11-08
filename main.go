package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/alehano/radiotbot/search"
	"github.com/alehano/radiotbot/shows"
	"github.com/blevesearch/bleve"

	"log"

	"net/http"

	"sort"

	"encoding/json"

	"fmt"

	"github.com/alehano/radiotbot/config"
	"gopkg.in/robfig/cron.v2"
)

const helpTextMd = "`Поиск!` - помощь\n`Поиск [запрос[:число результатов]]!` - поиск по выпускам\n`Выпуск [номер выпуска]!` - содержание выпуска\n\nВ запросе поддерживаются `-` и `+` префиксы и маска `*`\nПримеры: `Выпуск 520!`, `Поиск docker swarm!`, `Поиск +яндекс* +google :10!`\n"

var (
	searchIndex bleve.Index
	allShows    *shows.Shows
)

func main() {
	allShows = shows.Load()

	newSearchIndex, err := search.NewIndex()
	if err != nil {
		log.Fatal("Search index create error:", err)
	}
	searchIndex = newSearchIndex
	err = search.ReindexAll(searchIndex, allShows)
	if err != nil {
		log.Fatal("Search reindex error:", err)
	}

	// Update if last show was more than 7 days ago
	if allShows.Last().Date.Add(7 * 24 * time.Hour).Before(time.Now()) {
		update(allShows, searchIndex)
	}

	// Update every Monday at 6:00
	c := cron.New()
	_, err = c.AddFunc("0 6 * * * 1", func() {
		update(allShows, searchIndex)
	})
	if err != nil {
		log.Fatal("Add to cron error:", err)
	}

	// Run server
	log.Printf("%s\n", config.BotName)
	log.Printf("Total shows: %d, last show #%d\n", allShows.Len(), allShows.Last().ID)
	log.Printf("Bot running at %s\n", config.Port)

	http.HandleFunc("/event", panicRecover(webHandler))
	http.ListenAndServe(config.Port, nil)
}

func panicRecover(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC:%s\n", r)
			}
		}()
		f(w, r)
	}
}

func update(allShows *shows.Shows, index bleve.Index) {
	log.Println("Updating...")

	newShows := shows.Get(allShows.Last().ID, func(err error) {
		log.Println(err)
	})

	count := 0
	for _, show := range newShows.GetItems() {
		allShows.Add(show)
		err := search.AddToIndex(index, show)
		if err != nil {
			log.Println(err)
		}
		count++
	}

	sort.Sort(allShows)

	err := shows.Save(allShows)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%d show(s) updated\n", count)
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	type reqData struct {
		Text        string `json:"text"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}

	type respData struct {
		Text string `json:"text"`
		Bot  string `json:"bot"`
	}

	decoder := json.NewDecoder(r.Body)
	var rd reqData
	err := decoder.Decode(&rd)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}
	defer r.Body.Close()

	answer, err := query(strings.ToLower(rd.Text))
	if err != nil || answer == "" {
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}

	err = json.NewEncoder(w).Encode(respData{Text: answer, Bot: config.BotName})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}
}

func query(q string) (string, error) {
	if q == "поиск!" || q == "поиск !" {
		return getHelp()
	}

	if strings.HasPrefix(q, "поиск") && strings.HasSuffix(q, "!") {
		return search.Query(searchIndex, q, allShows)
	}

	if strings.HasPrefix(q, "выпуск") && strings.HasSuffix(q, "!") {
		return getShowDetail(q)
	}

	return "", nil
}

func getHelp() (string, error) {
	return helpTextMd, nil
}

func getShowDetail(q string) (string, error) {
	q = strings.Replace(q, "выпуск", "", 1)
	q = strings.Replace(q, "!", "", -1)
	q = strings.TrimSpace(q)
	num, err := strconv.Atoi(q)
	if err != nil {
		return "", nil
	}
	if show, ok := allShows.ItemsByID[num]; ok {
		out := fmt.Sprintf("**[Выпуск %d](%s)**\n", show.ID, show.URL)
		for _, topic := range show.TopicsMarkdown {
			out = fmt.Sprintf("%s* %s\n", out, topic)
		}
		if show.AudioURL != "" {
			out = fmt.Sprintf("%s[аудио](%s)", out, show.AudioURL)
		}
		if show.TorrentURL != "" {
			out = fmt.Sprintf("%s | [torrent](%s)", out, show.TorrentURL)
		}
		if show.ChatLogURL != "" {
			out = fmt.Sprintf("%s | [лог чата](%s)", out, show.ChatLogURL)
		}
		return out, nil
	}
	return "", nil
}
