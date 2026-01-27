package main

import (
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

	t.Run("loads config with required variables and default optionals", func(t *testing.T) {
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
		assert.Equal(t, "https://hnrss.org/frontpage", cfg.RSSURL)
	})

	t.Run("loads config with custom optional values", func(t *testing.T) {
		clearEnv()
		defer clearEnv()

		os.Setenv("LINE_ACCESS_TOKEN", "token123")
		os.Setenv("TARGET_USER_ID", "user123")
		os.Setenv("LINE_API_URL", "https://custom.api.line.me/push")
		os.Setenv("RSS_URL", "https://custom.rss.feed/news")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "token123", cfg.LineAccessToken)
		assert.Equal(t, "user123", cfg.TargetUserID)
		assert.Equal(t, "https://custom.api.line.me/push", cfg.LineAPIURL)
		assert.Equal(t, "https://custom.rss.feed/news", cfg.RSSURL)
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
		name          string
		token         string
		expectedMask  string
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
			name:         "two character token",
			token:        "ab",
			expectedMask: "***",
		},
		{
			name:         "three character token",
			token:        "abc",
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
			name:         "six character token",
			token:        "abcdef",
			expectedMask: "ab***ef",
		},
		{
			name:         "ten character token",
			token:        "abcdefghij",
			expectedMask: "ab***ij",
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
				RSSURL:          "https://rss.feed",
			}

			result := cfg.String()

			// Verify masked token is in the output
			assert.Contains(t, result, tt.expectedMask)
			// Verify original token is NOT in output (except for very short tokens where mask equals partial)
			if len(tt.token) > 4 {
				assert.NotContains(t, result, tt.token)
			}
			// Verify other fields are included
			assert.Contains(t, result, "user123")
			assert.Contains(t, result, "https://api.line.me")
			assert.Contains(t, result, "https://rss.feed")
		})
	}
}

func TestConfig_String_Format(t *testing.T) {
	cfg := &Config{
		LineAccessToken: "token12345",
		TargetUserID:    "userABC",
		LineAPIURL:      "https://api.example.com",
		RSSURL:          "https://rss.example.com",
	}

	result := cfg.String()

	// Verify the string contains expected field labels
	assert.Contains(t, result, "LineAPIURL")
	assert.Contains(t, result, "TargetUserID")
	assert.Contains(t, result, "RSSURL")
	assert.Contains(t, result, "LineAccessToken")

	// Verify values are quoted
	assert.Contains(t, result, `"userABC"`)
	assert.Contains(t, result, `"https://api.example.com"`)
	assert.Contains(t, result, `"https://rss.example.com"`)
	assert.Contains(t, result, `"to***45"`)
}
