package main

import (
	"fmt"
	"strings"
)

// FormatNews formats news items grouped by source into LINE messages.
// Each source becomes one LINE message. Order follows the feeds slice order.
func FormatNews(newsMap map[string][]NewsItem, feeds []FeedSource) []string {
	var messages []string

	for _, feed := range feeds {
		items, ok := newsMap[feed.Name]
		if !ok || len(items) == 0 {
			continue
		}

		var sb strings.Builder
		emoji := feed.Emoji
		if emoji == "" {
			emoji = "📰"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n━━━━━━━━━━━━━━", emoji, feed.Name))

		for i, item := range items {
			entry := fmt.Sprintf("\n\n%d. %s\n🔗 %s", i+1, item.Title, item.Link)

			// Check if adding this entry would exceed the LINE message limit
			if sb.Len()+len(entry) > MaxMessageLength {
				break
			}
			sb.WriteString(entry)
		}

		messages = append(messages, sb.String())
	}

	return messages
}
