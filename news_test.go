package main

import (
	"testing"
)

func TestGetHotNews(t *testing.T) {
	items, err := getNews()
	if err != nil {
		t.Fatalf("getHotNews() error = %v", err)
	}
	if len(items) == 0 {
		t.Error("getHotNews() returned no items")
	}

	// check the first item
	if items[0].Title == "" {
		t.Error("First item has empty title")
	}
	if items[0].Link == "" {
		t.Error("First item has empty link")
	}
}
