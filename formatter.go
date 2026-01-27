package main

import "fmt"

// FormatHackerNews converts HackerNews items to LINE message strings
func FormatHackerNews(items []Item) []string {
	messages := make([]string, len(items))
	for i, item := range items {
		messages[i] = fmt.Sprintf("%d. %s\n%s", i+1, item.Title, item.Link)
	}
	return messages
}
