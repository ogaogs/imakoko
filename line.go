package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// LINE API message character limit
const MaxMessageLength = 5000

// Maximum size for error response body
const maxErrorResponseSize = 4 * 1024 // 4KB

// MessageSender defines the capability to send messages
type MessageSender interface {
	Send(messages []string) error
}

// lineClient implements MessageSender using LINE API
type lineClient struct {
	httpClient   *http.Client
	apiURL       string
	accessToken  string
	targetUserID string
}

// NewLineClient creates a new LINE client
func NewLineClient(httpClient *http.Client, apiURL, accessToken, targetUserID string) MessageSender {
	return &lineClient{
		httpClient:   httpClient,
		apiURL:       apiURL,
		accessToken:  accessToken,
		targetUserID: targetUserID,
	}
}

// Send implements MessageSender interface for lineClient
func (c *lineClient) Send(messages []string) error {
	return sendLineMessage(c.httpClient, c.apiURL, c.accessToken, messages, c.targetUserID)
}

type LineMessages struct {
	SendTo   string        `json:"to"`
	Messages []LineContent `json:"messages"`
}

type LineContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func sendBatchLineMessage(sender MessageSender, messages []string) error {
	const batchSize = 5

	for i := 0; i < len(messages); i += batchSize {
		end := min(i+batchSize, len(messages))

		batch := messages[i:end]
		if err := sender.Send(batch); err != nil {
			return fmt.Errorf("failed to send batch %d-%d: %w", i+1, end, err)
		}
	}

	return nil
}

func sendLineMessage(httpClient *http.Client, apiURL string, accessToken string, messages []string, sendTo string) error {
	contents := make([]LineContent, len(messages))
	for i, msg := range messages {
		if len(msg) > MaxMessageLength {
			return fmt.Errorf("message %d exceeds LINE's %d character limit (has %d characters)", i+1, MaxMessageLength, len(msg))
		}
		contents[i] = LineContent{
			Type: "text",
			Text: msg,
		}
	}

	payload := LineMessages{
		SendTo:   sendTo,
		Messages: contents,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorResponseSize))
		return fmt.Errorf("LINE API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
