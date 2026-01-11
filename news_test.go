package main

import (
	"testing"
)

func TestFetchNews(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{name: "success",
			want: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := fetchNews()
			if (status) != tt.want {
				t.Errorf("fetchNews() = %v, want %v", status, tt.want)
			}
		})
	}
}
