package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	// Load configuration from environment variables
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Fetch news
	news, err := getNews(config.RSSURL)
	if err != nil {
		log.Fatalf("Failed to get news: %v", err)
	}

	// Format messages
	messages := FormatHackerNews(news)

	// Send LINE messages
	lineHTTPClient := &http.Client{Timeout: 30 * time.Second}
	client := NewLineClient(lineHTTPClient, config.LineAPIURL, config.LineAccessToken, config.TargetUserID)
	err = sendBatchLineMessage(client, messages)
	if err != nil {
		log.Fatalf("Failed to send LINE message: %v", err)
	}

	log.Println("Successfully sent messages")
}
