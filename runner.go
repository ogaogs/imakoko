package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// run executes a single news fetch and send cycle.
// Returns an error instead of calling log.Fatalf so callers can decide how to handle failures.
func run(config *Config) error {
	log.Println("Starting news fetch and send cycle")

	newsMap, err := getAllNews(config.Feeds)
	if err != nil {
		return fmt.Errorf("failed to get news: %w", err)
	}

	messages := FormatNews(newsMap, config.Feeds)

	lineHTTPClient := &http.Client{Timeout: 30 * time.Second}
	client := NewLineClient(lineHTTPClient, config.LineAPIURL, config.LineAccessToken, config.TargetUserID)
	if err := sendBatchLineMessage(client, messages); err != nil {
		return fmt.Errorf("failed to send LINE message: %w", err)
	}

	log.Println("Successfully sent messages")
	return nil
}
