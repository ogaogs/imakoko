package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRSS(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Article One</title>
      <link>https://example.com/1</link>
    </item>
    <item>
      <title>Article Two</title>
      <link>https://example.com/2</link>
    </item>
  </channel>
</rss>`)

	items, err := parseRSS(data)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Article One", items[0].Title)
	assert.Equal(t, "https://example.com/1", items[0].Link)
	assert.Equal(t, "Article Two", items[1].Title)
}

func TestParseRSS_InvalidXML(t *testing.T) {
	_, err := parseRSS([]byte(`not xml`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing RSS XML")
}

func TestParseAtom(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Atom Feed</title>
  <entry>
    <title>Atom Article</title>
    <link href="https://example.com/atom1" rel="alternate"/>
    <link href="https://example.com/atom1.json" rel="self"/>
  </entry>
  <entry>
    <title>Atom Article 2</title>
    <link href="https://example.com/atom2"/>
  </entry>
</feed>`)

	items, err := parseAtom(data)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Atom Article", items[0].Title)
	assert.Equal(t, "https://example.com/atom1", items[0].Link)
	assert.Equal(t, "https://example.com/atom2", items[1].Link)
}

func TestParseAtom_FallbackToFirstLink(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <title>Entry</title>
    <link href="https://example.com/only" rel="enclosure"/>
  </entry>
</feed>`)

	items, err := parseAtom(data)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "https://example.com/only", items[0].Link)
}

func TestParseReddit(t *testing.T) {
	data := []byte(`{
  "data": {
    "children": [
      {"data": {"title": "Reddit Post 1", "permalink": "/r/programming/comments/abc/post1/", "ups": 100}},
      {"data": {"title": "Reddit Post 2", "permalink": "/r/programming/comments/def/post2/", "ups": 50}}
    ]
  }
}`)

	items, err := parseReddit(data)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Reddit Post 1", items[0].Title)
	assert.Equal(t, "https://www.reddit.com/r/programming/comments/abc/post1/", items[0].Link)
}

func TestParseReddit_InvalidJSON(t *testing.T) {
	_, err := parseReddit([]byte(`not json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing Reddit JSON")
}

func TestParseFeed(t *testing.T) {
	t.Run("dispatches to RSS parser", func(t *testing.T) {
		data := []byte(`<rss><channel><item><title>T</title><link>L</link></item></channel></rss>`)
		items, err := parseFeed(data, FeedSource{Type: FeedTypeRSS})
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})

	t.Run("dispatches to Atom parser", func(t *testing.T) {
		data := []byte(`<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>T</title><link href="L"/></entry></feed>`)
		items, err := parseFeed(data, FeedSource{Type: FeedTypeAtom})
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})

	t.Run("dispatches to Reddit parser", func(t *testing.T) {
		data := []byte(`{"data":{"children":[{"data":{"title":"T","permalink":"/r/t/comments/1/t/"}}]}}`)
		items, err := parseFeed(data, FeedSource{Type: FeedTypeReddit})
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})

	t.Run("unknown feed type returns error", func(t *testing.T) {
		_, err := parseFeed([]byte{}, FeedSource{Type: "unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown feed type")
	})
}

func TestGetNews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<rss><channel>
			<item><title>A1</title><link>https://a.com/1</link></item>
			<item><title>A2</title><link>https://a.com/2</link></item>
			<item><title>A3</title><link>https://a.com/3</link></item>
		</channel></rss>`))
	}))
	defer server.Close()

	t.Run("fetches and parses successfully", func(t *testing.T) {
		source := FeedSource{Name: "Test", URL: server.URL, Type: FeedTypeRSS, MaxItems: 10}
		items, err := getNews(source)
		require.NoError(t, err)
		assert.Len(t, items, 3)
		assert.Equal(t, "Test", items[0].Source)
	})

	t.Run("limits items to MaxItems", func(t *testing.T) {
		source := FeedSource{Name: "Test", URL: server.URL, Type: FeedTypeRSS, MaxItems: 2}
		items, err := getNews(source)
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})
}

func TestGetAllNews(t *testing.T) {
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<rss><channel><item><title>RSS Item</title><link>https://rss.com</link></item></channel></rss>`))
	}))
	defer rssServer.Close()

	atomServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Atom Item</title><link href="https://atom.com"/></entry></feed>`))
	}))
	defer atomServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	t.Run("fetches from multiple sources", func(t *testing.T) {
		feeds := []FeedSource{
			{Name: "RSS Feed", URL: rssServer.URL, Type: FeedTypeRSS, MaxItems: 10},
			{Name: "Atom Feed", URL: atomServer.URL, Type: FeedTypeAtom, MaxItems: 10},
		}
		result, err := getAllNews(feeds)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Len(t, result["RSS Feed"], 1)
		assert.Len(t, result["Atom Feed"], 1)
	})

	t.Run("continues on partial failure", func(t *testing.T) {
		feeds := []FeedSource{
			{Name: "Good Feed", URL: rssServer.URL, Type: FeedTypeRSS, MaxItems: 10},
			{Name: "Bad Feed", URL: failServer.URL, Type: FeedTypeRSS, MaxItems: 10},
		}
		result, err := getAllNews(feeds)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result["Good Feed"], 1)
	})

	t.Run("returns error when all feeds fail", func(t *testing.T) {
		feeds := []FeedSource{
			{Name: "Bad 1", URL: failServer.URL, Type: FeedTypeRSS, MaxItems: 10},
			{Name: "Bad 2", URL: failServer.URL, Type: FeedTypeRSS, MaxItems: 10},
		}
		_, err := getAllNews(feeds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "all feeds failed")
	})
}
