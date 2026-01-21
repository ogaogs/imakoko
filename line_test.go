package main

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendLineMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"sentMessages": [
				{
				"id": "461230966842064897",
				"quoteToken": "IStG5h1Tz7b..."
				}
			]
		}`))
	}))
	defer server.Close()

	err := sendLineMessage(server.URL, "test-token", "Test Message")
	require.NoError(t, err)
}
