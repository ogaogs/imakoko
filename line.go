package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LineMessages struct {
	SendTo   string        `json:"to"`
	Messages []LineContent `json:"messages"`
}

type LineContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func sendLineMessage(apiURL string, accessToken string, messages []string, send_to string) error {

	contents := make([]LineContent, len(messages))
	for i, msg := range messages {
		contents[i] = LineContent{
			Type: "text",
			Text: msg,
		}
	}

	payload := LineMessages{
		SendTo:   send_to,
		Messages: contents,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: {%w}", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make NewRequest: {%w}", err)
	}

	req.Header.Set("Content-Type",
		"application/json")
	req.Header.Set("Authorization", "Bearer "+
		accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
