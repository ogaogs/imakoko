package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const maxResponseSize = 10 * 1024 * 1024 // 10MB limit

var feedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// NewsItem represents a single news article
type NewsItem struct {
	Title  string
	Link   string
	Source string
}

// RSS 2.0 structs
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

// Atom structs
type AtomFeed struct {
	Entries []AtomEntry `xml:"entry"`
}

type AtomEntry struct {
	Title string     `xml:"title"`
	Links []AtomLink `xml:"link"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// Reddit JSON structs
type RedditResponse struct {
	Data struct {
		Children []RedditChild `json:"children"`
	} `json:"data"`
}

type RedditChild struct {
	Data RedditPost `json:"data"`
}

type RedditPost struct {
	Title     string `json:"title"`
	Permalink string `json:"permalink"`
	Ups       int    `json:"ups"`
}

// fetchFeed fetches raw data from a feed source
func fetchFeed(source FeedSource) ([]byte, error) {
	req, err := http.NewRequest("GET", source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", source.Name, err)
	}

	// Reddit requires a User-Agent header
	if source.Type == FeedTypeReddit {
		req.Header.Set("User-Agent", "imakoko/1.0 (news aggregator)")
	}

	resp, err := feedHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", source.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status code: %d", source.Name, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", source.Name, err)
	}
	return body, nil
}

// parseRSS parses RSS 2.0 XML data into NewsItems
func parseRSS(data []byte) ([]NewsItem, error) {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("error parsing RSS XML: %w", err)
	}

	items := make([]NewsItem, len(rss.Channel.Items))
	for i, ri := range rss.Channel.Items {
		items[i] = NewsItem{Title: ri.Title, Link: ri.Link}
	}
	return items, nil
}

// parseAtom parses Atom XML data into NewsItems
func parseAtom(data []byte) ([]NewsItem, error) {
	var feed AtomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("error parsing Atom XML: %w", err)
	}

	items := make([]NewsItem, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		link := ""
		for _, l := range entry.Links {
			if l.Rel == "alternate" || l.Rel == "" {
				link = l.Href
				break
			}
		}
		// Fallback to first link if no alternate found
		if link == "" && len(entry.Links) > 0 {
			link = entry.Links[0].Href
		}
		items = append(items, NewsItem{Title: entry.Title, Link: link})
	}
	return items, nil
}

// parseReddit parses Reddit JSON API response into NewsItems
func parseReddit(data []byte) ([]NewsItem, error) {
	var resp RedditResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("error parsing Reddit JSON: %w", err)
	}

	items := make([]NewsItem, 0, len(resp.Data.Children))
	for _, child := range resp.Data.Children {
		link := "https://www.reddit.com" + child.Data.Permalink
		items = append(items, NewsItem{Title: child.Data.Title, Link: link})
	}
	return items, nil
}

// parseFeed dispatches parsing based on feed type
func parseFeed(data []byte, source FeedSource) ([]NewsItem, error) {
	switch source.Type {
	case FeedTypeRSS:
		return parseRSS(data)
	case FeedTypeAtom:
		return parseAtom(data)
	case FeedTypeReddit:
		return parseReddit(data)
	default:
		return nil, fmt.Errorf("unknown feed type: %s", source.Type)
	}
}

// getNews fetches and parses a single feed source
func getNews(source FeedSource) ([]NewsItem, error) {
	data, err := fetchFeed(source)
	if err != nil {
		return nil, err
	}
	items, err := parseFeed(data, source)
	if err != nil {
		return nil, err
	}

	// Set source name and limit items
	for i := range items {
		items[i].Source = source.Name
	}
	if source.MaxItems > 0 && len(items) > source.MaxItems {
		items = items[:source.MaxItems]
	}

	return items, nil
}

// getAllNews fetches news from all feed sources.
// Returns a map of source name to items. Logs and skips individual feed failures.
// Returns error only if ALL feeds fail.
func getAllNews(feeds []FeedSource) (map[string][]NewsItem, error) {
	result := make(map[string][]NewsItem)
	var lastErr error

	for _, feed := range feeds {
		items, err := getNews(feed)
		if err != nil {
			log.Printf("Warning: failed to fetch %s: %v", feed.Name, err)
			lastErr = err
			continue
		}
		if len(items) > 0 {
			result[feed.Name] = items
		}
	}

	if len(result) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("all feeds failed, last error: %w", lastErr)
		}
		return nil, fmt.Errorf("no feeds configured")
	}

	return result, nil
}
