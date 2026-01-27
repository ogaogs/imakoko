package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxResponseSize = 10 * 1024 * 1024 // 10MB limit

var rssHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

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

func fetchHNRSS(rssURL string) ([]byte, error) {
	resp, err := rssHTTPClient.Get(rssURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

func parseNews(data []byte) ([]Item, error) {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("error parsing XML: %w", err)
	}

	return rss.Channel.Items, nil
}

func getNews(rssURL string) ([]Item, error) {
	data, err := fetchHNRSS(rssURL)
	if err != nil {
		return nil, err
	}
	items, err := parseNews(data)
	if err != nil {
		return nil, err
	}
	return items, nil
}
