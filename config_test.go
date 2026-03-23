package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Helper to clear and restore environment variables
	clearEnv := func() {
		os.Unsetenv("LINE_ACCESS_TOKEN")
		os.Unsetenv("TARGET_USER_ID")
		os.Unsetenv("LINE_API_URL")
		os.Unsetenv("RSS_URL")
		os.Unsetenv("FEED_SOURCES")
		os.Unsetenv("ENABLED_FEEDS")
	}

	t.Run("returns error when LINE_ACCESS_TOKEN is missing", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("TARGET_USER_ID", "user123")

		cfg, err := LoadConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "LINE_ACCESS_TOKEN environment variable is required")
	})

	t.Run("returns error when TARGET_USER_ID is missing", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")

		cfg, err := LoadConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "TARGET_USER_ID environment variable is required")
	})

	t.Run("loads config with default enabled feeds", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "token123", cfg.LineAccessToken)
		assert.Equal(t, "user123", cfg.TargetUserID)
		assert.Equal(t, "https://api.line.me/v2/bot/message/push", cfg.LineAPIURL)
		assert.Len(t, cfg.Feeds, 3)
		assert.Equal(t, "Lobsters", cfg.Feeds[0].Name)
		assert.Equal(t, "Zenn Trending", cfg.Feeds[1].Name)
		assert.Equal(t, "Hacker News", cfg.Feeds[2].Name)
	})

	t.Run("ENABLED_FEEDS overrides default enabled feeds", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")
		os.Setenv("ENABLED_FEEDS", "Publickey, TechCrunch")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Len(t, cfg.Feeds, 2)
		assert.Equal(t, "Publickey", cfg.Feeds[0].Name)
		assert.Equal(t, "TechCrunch", cfg.Feeds[1].Name)
	})

	t.Run("RSS_URL backward compatibility", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")
		os.Setenv("RSS_URL", "https://custom.rss.feed/news")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Len(t, cfg.Feeds, 1)
		assert.Equal(t, "News", cfg.Feeds[0].Name)
		assert.Equal(t, "https://custom.rss.feed/news", cfg.Feeds[0].URL)
		assert.Equal(t, FeedTypeRSS, cfg.Feeds[0].Type)
	})

	t.Run("FEED_SOURCES JSON overrides defaults", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")
		os.Setenv("FEED_SOURCES", `[{"name":"HN","url":"https://hnrss.org/frontpage","type":"rss","max_items":5}]`)

		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Len(t, cfg.Feeds, 1)
		assert.Equal(t, "HN", cfg.Feeds[0].Name)
		assert.Equal(t, 5, cfg.Feeds[0].MaxItems)
	})

	t.Run("FEED_SOURCES takes priority over RSS_URL", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")
		os.Setenv("RSS_URL", "https://should.be.ignored")
		os.Setenv("FEED_SOURCES", `[{"name":"Custom","url":"https://custom.feed","type":"atom","max_items":3}]`)

		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Len(t, cfg.Feeds, 1)
		assert.Equal(t, "Custom", cfg.Feeds[0].Name)
		assert.Equal(t, FeedTypeAtom, cfg.Feeds[0].Type)
	})

	t.Run("invalid FEED_SOURCES JSON returns error", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")
		os.Setenv("FEED_SOURCES", `invalid json`)

		cfg, err := LoadConfig()
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "invalid FEED_SOURCES JSON")
	})

	t.Run("returns error when both required variables are missing", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		cfg, err := LoadConfig()
		assert.Nil(t, cfg)
		// LINE_ACCESS_TOKEN is checked first
		assert.EqualError(t, err, "LINE_ACCESS_TOKEN environment variable is required")
	})
}

func TestConfig_String(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		expectedMask string
	}{
		{
			name:         "empty token",
			token:        "",
			expectedMask: "***",
		},
		{
			name:         "single character token",
			token:        "a",
			expectedMask: "***",
		},
		{
			name:         "four character token (boundary)",
			token:        "abcd",
			expectedMask: "***",
		},
		{
			name:         "five character token (boundary)",
			token:        "abcde",
			expectedMask: "ab***de",
		},
		{
			name:         "long realistic token",
			token:        "U9WodLAjAvdbCLCnraoA8imAuaee6yQs5N2e6kodYPpkRqa2zQdt0RKPAYcMNu",
			expectedMask: "U9***Nu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LineAccessToken: tt.token,
				TargetUserID:    "user123",
				LineAPIURL:      "https://api.line.me",
				Feeds:           defaultFeeds(),
			}

			result := cfg.String()

			assert.Contains(t, result, tt.expectedMask)
			if len(tt.token) > 4 {
				assert.NotContains(t, result, tt.token)
			}
			assert.Contains(t, result, "user123")
			assert.Contains(t, result, "https://api.line.me")
			assert.Contains(t, result, fmt.Sprintf("%d sources", len(defaultFeeds())))
		})
	}
}

func TestConfig_String_Format(t *testing.T) {
	cfg := &Config{
		LineAccessToken: "token12345",
		TargetUserID:    "userABC",
		LineAPIURL:      "https://api.example.com",
		Feeds:           []FeedSource{{Name: "Test", URL: "https://test.com", Type: FeedTypeRSS, MaxItems: 5}},
	}

	result := cfg.String()

	assert.Contains(t, result, "LineAPIURL")
	assert.Contains(t, result, "TargetUserID")
	assert.Contains(t, result, "Feeds")
	assert.Contains(t, result, "LineAccessToken")
	assert.Contains(t, result, `"userABC"`)
	assert.Contains(t, result, `"https://api.example.com"`)
	assert.Contains(t, result, "1 sources")
	assert.Contains(t, result, `"to***45"`)
}

func TestDefaultFeeds(t *testing.T) {
	feeds := defaultFeeds()

	assert.Len(t, feeds, 9)

	// Verify feed types
	rssCount, atomCount, redditCount := 0, 0, 0
	for _, f := range feeds {
		switch f.Type {
		case FeedTypeRSS:
			rssCount++
		case FeedTypeAtom:
			atomCount++
		case FeedTypeReddit:
			redditCount++
		}
		assert.NotEmpty(t, f.Name)
		assert.NotEmpty(t, f.URL)
		assert.Greater(t, f.MaxItems, 0)
	}
	assert.Equal(t, 5, rssCount)
	assert.Equal(t, 2, atomCount)
	assert.Equal(t, 2, redditCount)
}

func TestFilterFeedsByNames(t *testing.T) {
	feeds := defaultFeeds()

	t.Run("filters to specified names", func(t *testing.T) {
		filtered := filterFeedsByNames(feeds, []string{"Hacker News", "Lobsters"})
		assert.Len(t, filtered, 2)
		assert.Equal(t, "Hacker News", filtered[0].Name)
		assert.Equal(t, "Lobsters", filtered[1].Name)
	})

	t.Run("preserves enabledNames order", func(t *testing.T) {
		filtered := filterFeedsByNames(feeds, []string{"Lobsters", "Hacker News"})
		assert.Len(t, filtered, 2)
		assert.Equal(t, "Lobsters", filtered[0].Name)
		assert.Equal(t, "Hacker News", filtered[1].Name)
	})

	t.Run("ignores unknown names", func(t *testing.T) {
		filtered := filterFeedsByNames(feeds, []string{"NonExistent", "Hacker News"})
		assert.Len(t, filtered, 1)
		assert.Equal(t, "Hacker News", filtered[0].Name)
	})

	t.Run("returns empty slice for no matches", func(t *testing.T) {
		filtered := filterFeedsByNames(feeds, []string{"NonExistent"})
		assert.Empty(t, filtered)
	})
}

func TestDefaultEnabledFeedNames(t *testing.T) {
	names := defaultEnabledFeedNames()
	assert.Equal(t, []string{"Lobsters", "Zenn Trending", "Hacker News"}, names)
}

func TestLoadEnabledFeedNames(t *testing.T) {
	t.Run("returns nil when ENABLED_FEEDS is not set", func(t *testing.T) {
		os.Unsetenv("ENABLED_FEEDS")
		assert.Nil(t, loadEnabledFeedNames())
	})

	t.Run("parses comma-separated names", func(t *testing.T) {
		os.Setenv("ENABLED_FEEDS", "Hacker News, Publickey ,TechCrunch")
		defer os.Unsetenv("ENABLED_FEEDS")

		names := loadEnabledFeedNames()
		assert.Equal(t, []string{"Hacker News", "Publickey", "TechCrunch"}, names)
	})

	t.Run("skips empty entries", func(t *testing.T) {
		os.Setenv("ENABLED_FEEDS", "Hacker News,,Lobsters,")
		defer os.Unsetenv("ENABLED_FEEDS")

		names := loadEnabledFeedNames()
		assert.Equal(t, []string{"Hacker News", "Lobsters"}, names)
	})
}
