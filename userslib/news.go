package userslib

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type NewsItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

func GetXakerNewsHandler(w http.ResponseWriter, r *http.Request) {
	news, err := GetXakerNews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(news)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func GetXakerNews() ([]NewsItem, error) {
	var result []NewsItem

	url := "https://xakep.ru/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	doc.Find(".block-article-content-wrapper > header > h3 > a").Each(func(index int, sel *goquery.Selection) {
		if index < 10 {
			resultObject := NewsItem{
				ID:    index + 1,
				Title: strings.TrimSpace(sel.Text()),
				Link:  sel.AttrOr("href", ""),
			}
			result = append(result, resultObject)
		}
	})

	return result, nil
}
