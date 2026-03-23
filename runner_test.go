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

const testAtomResponse = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <title>Atom Article</title>
    <link href="https://example.com/atom1" rel="alternate"/>
  </entry>
</feed>`

func TestRun_Success(t *testing.T) {
	// Mock RSS server
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRSSResponse))
	}))
	defer rssServer.Close()

	// Mock Atom server
	atomServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testAtomResponse))
	}))
	defer atomServer.Close()

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
		Feeds: []FeedSource{
			{Name: "RSS Feed", URL: rssServer.URL, Type: FeedTypeRSS, MaxItems: 10},
			{Name: "Atom Feed", URL: atomServer.URL, Type: FeedTypeAtom, MaxItems: 10},
		},
	}

	err := run(config)
	require.NoError(t, err)
}

func TestRun_RSSFetchError(t *testing.T) {
	// All feeds return errors
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	config := &Config{
		LineAccessToken: "test-token",
		TargetUserID:    "test-user",
		LineAPIURL:      "http://localhost:0",
		Feeds: []FeedSource{
			{Name: "Bad Feed", URL: failServer.URL, Type: FeedTypeRSS, MaxItems: 10},
		},
	}

	err := run(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get news")
}

func TestRun_PartialFeedFailure(t *testing.T) {
	// One good feed, one bad
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRSSResponse))
	}))
	defer rssServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	lineServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer lineServer.Close()

	config := &Config{
		LineAccessToken: "test-token",
		TargetUserID:    "test-user",
		LineAPIURL:      lineServer.URL,
		Feeds: []FeedSource{
			{Name: "Good Feed", URL: rssServer.URL, Type: FeedTypeRSS, MaxItems: 10},
			{Name: "Bad Feed", URL: failServer.URL, Type: FeedTypeRSS, MaxItems: 10},
		},
	}

	err := run(config)
	require.NoError(t, err)
}

func TestRun_LineSendError(t *testing.T) {
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testRSSResponse))
	}))
	defer rssServer.Close()

	lineServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid token"}`))
	}))
	defer lineServer.Close()

	config := &Config{
		LineAccessToken: "invalid-token",
		TargetUserID:    "test-user",
		LineAPIURL:      lineServer.URL,
		Feeds: []FeedSource{
			{Name: "Test Feed", URL: rssServer.URL, Type: FeedTypeRSS, MaxItems: 10},
		},
	}

	err := run(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send LINE message")
}
