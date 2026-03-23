package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// FeedType represents the type of a feed source
type FeedType string

const (
	FeedTypeRSS    FeedType = "rss"
	FeedTypeAtom   FeedType = "atom"
	FeedTypeReddit FeedType = "reddit"
)

// FeedSource defines a news feed to fetch
type FeedSource struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Type     FeedType `json:"type"`
	MaxItems int      `json:"max_items"`
	Emoji    string   `json:"emoji,omitempty"`
}

// Config holds application settings
type Config struct {
	LineAccessToken string
	TargetUserID    string
	LineAPIURL      string
	Feeds           []FeedSource
}

// defaultFeeds returns the built-in feed sources
func defaultFeeds() []FeedSource {
	return []FeedSource{
		{Name: "はてなブックマーク IT", URL: "https://b.hatena.ne.jp/hotentry/it.rss", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🔖"},
		{Name: "Hacker News", URL: "https://hnrss.org/frontpage", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🟧"},
		{Name: "Zenn Trending", URL: "https://zenn.dev/feed", Type: FeedTypeRSS, MaxItems: 10, Emoji: "✍️"},
		{Name: "Lobsters", URL: "https://lobste.rs/rss", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🦞"},
		{Name: "Publickey", URL: "https://www.publickey1.jp/atom.xml", Type: FeedTypeAtom, MaxItems: 10, Emoji: "🔑"},
		{Name: "TechCrunch", URL: "https://techcrunch.com/feed/", Type: FeedTypeRSS, MaxItems: 10, Emoji: "💥"},
		{Name: "The Verge", URL: "https://www.theverge.com/rss/index.xml", Type: FeedTypeAtom, MaxItems: 10, Emoji: "📱"},
		{Name: "r/programming", URL: "https://old.reddit.com/r/programming/hot.json?t=day&limit=10", Type: FeedTypeReddit, MaxItems: 10, Emoji: "💻"},
		{Name: "r/technology", URL: "https://old.reddit.com/r/technology/hot.json?t=day&limit=10", Type: FeedTypeReddit, MaxItems: 10, Emoji: "🤖"},
	}
}

// LoadConfig reads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Required environment variables
	accessToken := os.Getenv("LINE_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("LINE_ACCESS_TOKEN environment variable is required")
	}

	targetUserID := os.Getenv("TARGET_USER_ID")
	if targetUserID == "" {
		return nil, errors.New("TARGET_USER_ID environment variable is required")
	}

	// Optional environment variables with defaults
	apiURL := os.Getenv("LINE_API_URL")
	if apiURL == "" {
		apiURL = "https://api.line.me/v2/bot/message/push"
	}

	// Feed sources: FEED_SOURCES > RSS_URL > defaults
	feeds, err := loadFeeds()
	if err != nil {
		return nil, fmt.Errorf("failed to load feed sources: %w", err)
	}

	return &Config{
		LineAccessToken: accessToken,
		TargetUserID:    targetUserID,
		LineAPIURL:      apiURL,
		Feeds:           feeds,
	}, nil
}

// loadFeeds determines feed sources from environment variables
func loadFeeds() ([]FeedSource, error) {
	// Priority 1: FEED_SOURCES JSON
	if feedJSON := os.Getenv("FEED_SOURCES"); feedJSON != "" {
		var feeds []FeedSource
		if err := json.Unmarshal([]byte(feedJSON), &feeds); err != nil {
			return nil, fmt.Errorf("invalid FEED_SOURCES JSON: %w", err)
		}
		return feeds, nil
	}

	// Priority 2: RSS_URL (backward compatibility)
	if rssURL := os.Getenv("RSS_URL"); rssURL != "" {
		return []FeedSource{
			{Name: "News", URL: rssURL, Type: FeedTypeRSS, MaxItems: 30},
		}, nil
	}

	// Priority 3: default feeds
	return defaultFeeds(), nil
}

// String returns the string representation of the config (token is masked)
func (c *Config) String() string {
	maskedToken := "***"
	if len(c.LineAccessToken) > 4 {
		maskedToken = c.LineAccessToken[:2] + "***" + c.LineAccessToken[len(c.LineAccessToken)-2:]
	}
	return fmt.Sprintf("Config{LineAPIURL: %q, TargetUserID: %q, Feeds: %d sources, LineAccessToken: %q}",
		c.LineAPIURL, c.TargetUserID, len(c.Feeds), maskedToken)
}
