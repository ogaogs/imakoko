package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRSSResponse = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Hacker News</title>
    <item>
      <title>Test Article 1</title>
      <link>https://example.com/1</link>
    </item>
    <item>
      <title>Test Article 2</title>
      <link>https://example.com/2</link>
    </item>
  </channel>
</rss>`

func TestRun_Success(t *testing.T) {
	// Mock RSS server
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRSSResponse))
	}))
	defer rssServer.Close()

	// Mock LINE API server
	lineServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sentMessages": []}`))
	}))
	defer lineServer.Close()

	config := &Config{
		LineAccessToken: "test-token",
		TargetUserID:    "test-user",
		LineAPIURL:      lineServer.URL,
		RSSURL:          rssServer.URL,
	}

	err := run(config)
	require.NoError(t, err)
}

func TestRun_RSSFetchError(t *testing.T) {
	// Use a server that returns an error
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer rssServer.Close()

	config := &Config{
		LineAccessToken: "test-token",
		TargetUserID:    "test-user",
		LineAPIURL:      "http://localhost:0",
		RSSURL:          rssServer.URL,
	}

	err := run(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get news")
}

func TestRun_LineSendError(t *testing.T) {
	// Mock RSS server that returns valid data
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRSSResponse))
	}))
	defer rssServer.Close()

	// Mock LINE API server that returns an error
	lineServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid token"}`))
	}))
	defer lineServer.Close()

	config := &Config{
		LineAccessToken: "invalid-token",
		TargetUserID:    "test-user",
		LineAPIURL:      lineServer.URL,
		RSSURL:          rssServer.URL,
	}

	err := run(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send LINE message")
}
