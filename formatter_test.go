package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatNews(t *testing.T) {
	t.Run("single source formatting", func(t *testing.T) {
		newsMap := map[string][]NewsItem{
			"Hacker News": {
				{Title: "Article One", Link: "https://example.com/1", Source: "Hacker News"},
				{Title: "Article Two", Link: "https://example.com/2", Source: "Hacker News"},
			},
		}
		feeds := []FeedSource{{Name: "Hacker News", URL: "https://hn.com", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🟧"}}

		result := FormatNews(newsMap, feeds)

		require.Len(t, result, 1)
		assert.Contains(t, result[0], "🟧 Hacker News")
		assert.Contains(t, result[0], "1. Article One")
		assert.Contains(t, result[0], "🔗 https://example.com/1")
		assert.Contains(t, result[0], "2. Article Two")
	})

	t.Run("multiple sources in feed order", func(t *testing.T) {
		newsMap := map[string][]NewsItem{
			"Source B": {{Title: "B1", Link: "https://b.com", Source: "Source B"}},
			"Source A": {{Title: "A1", Link: "https://a.com", Source: "Source A"}},
		}
		feeds := []FeedSource{
			{Name: "Source A", URL: "https://a.com", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🅰️"},
			{Name: "Source B", URL: "https://b.com", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🅱️"},
		}

		result := FormatNews(newsMap, feeds)

		require.Len(t, result, 2)
		assert.Contains(t, result[0], "🅰️ Source A")
		assert.Contains(t, result[1], "🅱️ Source B")
	})

	t.Run("skips sources with no items", func(t *testing.T) {
		newsMap := map[string][]NewsItem{
			"Has Items": {{Title: "Item", Link: "https://item.com", Source: "Has Items"}},
		}
		feeds := []FeedSource{
			{Name: "No Items", URL: "https://empty.com", Type: FeedTypeRSS, MaxItems: 10},
			{Name: "Has Items", URL: "https://item.com", Type: FeedTypeRSS, MaxItems: 10, Emoji: "✅"},
		}

		result := FormatNews(newsMap, feeds)

		require.Len(t, result, 1)
		assert.Contains(t, result[0], "✅ Has Items")
	})

	t.Run("empty news map returns empty slice", func(t *testing.T) {
		newsMap := map[string][]NewsItem{}
		feeds := []FeedSource{{Name: "Empty", URL: "https://e.com", Type: FeedTypeRSS, MaxItems: 10}}

		result := FormatNews(newsMap, feeds)

		assert.Empty(t, result)
	})

	t.Run("truncates at MaxMessageLength", func(t *testing.T) {
		items := make([]NewsItem, 100)
		for i := range items {
			items[i] = NewsItem{
				Title:  strings.Repeat("A", 100),
				Link:   "https://example.com/" + strings.Repeat("x", 80),
				Source: "Big Feed",
			}
		}
		newsMap := map[string][]NewsItem{"Big Feed": items}
		feeds := []FeedSource{{Name: "Big Feed", URL: "https://big.com", Type: FeedTypeRSS, MaxItems: 100}}

		result := FormatNews(newsMap, feeds)

		require.Len(t, result, 1)
		assert.LessOrEqual(t, len(result[0]), MaxMessageLength)
	})

	t.Run("format includes proper spacing", func(t *testing.T) {
		newsMap := map[string][]NewsItem{
			"Test": {
				{Title: "First", Link: "https://first.com", Source: "Test"},
				{Title: "Second", Link: "https://second.com", Source: "Test"},
			},
		}
		feeds := []FeedSource{{Name: "Test", URL: "https://t.com", Type: FeedTypeRSS, MaxItems: 10, Emoji: "🧪"}}

		result := FormatNews(newsMap, feeds)

		require.Len(t, result, 1)
		expected := "🧪 Test\n━━━━━━━━━━━━━━\n\n1. First\n🔗 https://first.com\n\n2. Second\n🔗 https://second.com"
		assert.Equal(t, expected, result[0])
	})

	t.Run("fallback emoji when not set", func(t *testing.T) {
		newsMap := map[string][]NewsItem{
			"No Emoji": {{Title: "Item", Link: "https://item.com", Source: "No Emoji"}},
		}
		feeds := []FeedSource{{Name: "No Emoji", URL: "https://e.com", Type: FeedTypeRSS, MaxItems: 10}}

		result := FormatNews(newsMap, feeds)

		require.Len(t, result, 1)
		assert.Contains(t, result[0], "📰 No Emoji")
	})
}
