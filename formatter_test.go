package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatHackerNews(t *testing.T) {
	t.Run("empty items returns empty slice", func(t *testing.T) {
		items := []Item{}
		result := FormatHackerNews(items)

		assert.Empty(t, result)
		assert.NotNil(t, result)
	})

	t.Run("single item formatting", func(t *testing.T) {
		items := []Item{
			{Title: "Test Article", Link: "https://example.com/article"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. Test Article\nhttps://example.com/article", result[0])
	})

	t.Run("multiple items with correct numbering", func(t *testing.T) {
		items := []Item{
			{Title: "First Article", Link: "https://example.com/1"},
			{Title: "Second Article", Link: "https://example.com/2"},
			{Title: "Third Article", Link: "https://example.com/3"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 3)
		assert.Equal(t, "1. First Article\nhttps://example.com/1", result[0])
		assert.Equal(t, "2. Second Article\nhttps://example.com/2", result[1])
		assert.Equal(t, "3. Third Article\nhttps://example.com/3", result[2])
	})

	t.Run("item with empty title", func(t *testing.T) {
		items := []Item{
			{Title: "", Link: "https://example.com/article"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. \nhttps://example.com/article", result[0])
	})

	t.Run("item with empty link", func(t *testing.T) {
		items := []Item{
			{Title: "Test Article", Link: ""},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. Test Article\n", result[0])
	})

	t.Run("item with both empty title and link", func(t *testing.T) {
		items := []Item{
			{Title: "", Link: ""},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. \n", result[0])
	})

	t.Run("title with special characters", func(t *testing.T) {
		items := []Item{
			{Title: "Test <script>alert('xss')</script> Article", Link: "https://example.com/article"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. Test <script>alert('xss')</script> Article\nhttps://example.com/article", result[0])
	})

	t.Run("title with unicode characters", func(t *testing.T) {
		items := []Item{
			{Title: "日本語タイトル 🚀", Link: "https://example.com/jp"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. 日本語タイトル 🚀\nhttps://example.com/jp", result[0])
	})

	t.Run("URL with query parameters and fragments", func(t *testing.T) {
		items := []Item{
			{Title: "Article", Link: "https://example.com/article?id=123&ref=hn#section"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. Article\nhttps://example.com/article?id=123&ref=hn#section", result[0])
	})

	t.Run("title with newlines preserved", func(t *testing.T) {
		items := []Item{
			{Title: "Title\nWith\nNewlines", Link: "https://example.com"},
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 1)
		assert.Equal(t, "1. Title\nWith\nNewlines\nhttps://example.com", result[0])
	})

	t.Run("large number of items maintains correct numbering", func(t *testing.T) {
		items := make([]Item, 100)
		for i := 0; i < 100; i++ {
			items[i] = Item{Title: "Article", Link: "https://example.com"}
		}

		result := FormatHackerNews(items)

		assert.Len(t, result, 100)
		assert.Contains(t, result[0], "1. Article")
		assert.Contains(t, result[9], "10. Article")
		assert.Contains(t, result[99], "100. Article")
	})
}
