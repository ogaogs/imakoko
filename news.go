package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

const RSSURL = "https://hnrss.org/frontpage"

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

func fetchHNRSS() ([]byte, error) {
	resp, err := http.Get(RSSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to response body: %w", err)
	}
	return body, nil
}

func parseNews(data []byte) ([]Item, error) {
	var rss RSS
	var items []Item
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("Error parsing XML: %w", err)
	}

	for _, item := range rss.Channel.Items {
		items = append(items, item)
	}
	return items, nil
}

func getNews() ([]Item, error) {
	data, err := fetchHNRSS()
	if err != nil {
		return nil, err
	}
	items, err := parseNews(data)
	if err != nil {
		return nil, err
	}
	return items, nil
}
