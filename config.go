package main

import (
	"errors"
	"fmt"
	"os"
)

// Config はアプリケーション設定を保持する
type Config struct {
	LineAccessToken string
	TargetUserID    string
	LineAPIURL      string
	RSSURL          string
}

// LoadConfig は環境変数から設定を読み込む
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

	rssURL := os.Getenv("RSS_URL")
	if rssURL == "" {
		rssURL = "https://hnrss.org/frontpage"
	}

	return &Config{
		LineAccessToken: accessToken,
		TargetUserID:    targetUserID,
		LineAPIURL:      apiURL,
		RSSURL:          rssURL,
	}, nil
}

// String は設定の文字列表現を返す（トークンはマスキング）
func (c *Config) String() string {
	maskedToken := "***"
	if len(c.LineAccessToken) > 4 {
		maskedToken = c.LineAccessToken[:2] + "***" + c.LineAccessToken[len(c.LineAccessToken)-2:]
	}
	return fmt.Sprintf("Config{LineAPIURL: %q, TargetUserID: %q, RSSURL: %q, LineAccessToken: %q}",
		c.LineAPIURL, c.TargetUserID, c.RSSURL, maskedToken)
}
