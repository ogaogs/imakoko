package main

import (
	"fmt"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const noErrorCall = -1 // Constant to indicate no error should occur

// mockMessageSender is a mock implementation for testing
type mockMessageSender struct {
	sentBatches [][]string
	err         error
	callCount   int
	errorOnCall int // Return error on this call number (0-indexed, use noErrorCall for no error)
}

func (m *mockMessageSender) Send(messages []string) error {
	if m.errorOnCall != noErrorCall && m.callCount == m.errorOnCall {
		m.callCount++
		return m.err
	}

	m.sentBatches = append(m.sentBatches, messages)
	m.callCount++
	return nil
}

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

	messages := []string{"Test Message"}
	err := sendLineMessage(http.DefaultClient, server.URL, "test-token", messages, "test-user-id")
	require.NoError(t, err)
}

func TestSendBatchLineMessage(t *testing.T) {
	tests := []struct {
		name          string
		messageCount  int
		expectedCalls int
		shouldError   bool
		errorOnCall   int // 0-indexed
	}{
		{
			name:          "1 message - single batch",
			messageCount:  1,
			expectedCalls: 1,
		},
		{
			name:          "5 messages - single batch",
			messageCount:  5,
			expectedCalls: 1,
		},
		{
			name:          "6 messages - two batches",
			messageCount:  6,
			expectedCalls: 2,
		},
		{
			name:          "10 messages - two batches",
			messageCount:  10,
			expectedCalls: 2,
		},
		{
			name:          "12 messages - three batches",
			messageCount:  12,
			expectedCalls: 3,
		},
		{
			name:          "error on second batch",
			messageCount:  7,
			expectedCalls: 2,
			shouldError:   true,
			errorOnCall:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock instead of httptest
			mock := &mockMessageSender{
				errorOnCall: noErrorCall, // Default: no error
			}

			if tt.shouldError {
				mock.err = assert.AnError
				mock.errorOnCall = tt.errorOnCall
			}

			// Create test messages
			messages := make([]string, tt.messageCount)
			for i := 0; i < tt.messageCount; i++ {
				messages[i] = fmt.Sprintf("Message %d", i+1)
			}

			// Execute
			err := sendBatchLineMessage(mock, messages)

			// Assertions
			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to send batch")
			} else {
				require.NoError(t, err)
				// Verify batch contents
				assert.Equal(t, tt.expectedCalls, len(mock.sentBatches), "unexpected number of batches")

				// Verify each batch size
				totalMessages := 0
				for i, batch := range mock.sentBatches {
					if i < len(mock.sentBatches)-1 {
						// All batches except the last should have 5 messages
						assert.Equal(t, 5, len(batch), "batch %d should have 5 messages", i)
					}
					totalMessages += len(batch)
				}
				assert.Equal(t, tt.messageCount, totalMessages, "total messages should match")
			}
		})
	}
}

func TestSendBatchLineMessage_EmptyMessages(t *testing.T) {
	mock := &mockMessageSender{}

	err := sendBatchLineMessage(mock, []string{})

	require.NoError(t, err)
	assert.Equal(t, 0, len(mock.sentBatches), "should not send with empty messages")
}

func TestNewLineClient(t *testing.T) {
	tests := []struct {
		name         string
		apiURL       string
		accessToken  string
		targetUserID string
	}{
		{
			name:         "creates client with valid parameters",
			apiURL:       "https://api.line.me/v2/bot/message/push",
			accessToken:  "test-token-123",
			targetUserID: "U1234567890abcdef",
		},
		{
			name:         "creates client with empty strings",
			apiURL:       "",
			accessToken:  "",
			targetUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewLineClient(http.DefaultClient, tt.apiURL, tt.accessToken, tt.targetUserID)

			require.NotNil(t, client)
			// Verify it implements MessageSender interface
			var _ MessageSender = client
		})
	}
}

func TestLineClient_Send(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sentMessages": [{"id": "123", "quoteToken": "token"}]}`))
	}))
	defer server.Close()

	client := NewLineClient(http.DefaultClient, server.URL, "test-token", "U123456")

	tests := []struct {
		name     string
		messages []string
	}{
		{
			name:     "sends single message",
			messages: []string{"Hello"},
		},
		{
			name:     "sends multiple messages",
			messages: []string{"Hello", "World", "Test"},
		},
		{
			name:     "sends empty array",
			messages: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Send(tt.messages)
			require.NoError(t, err)
		})
	}

	// Verify server was called correct number of times (3 test cases)
	assert.Equal(t, 3, callCount)
}

func TestSendLineMessage_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		serverResponse string
		expectError    string
	}{
		{
			name:           "handles 400 Bad Request",
			statusCode:     http.StatusBadRequest,
			serverResponse: `{"message": "Invalid request"}`,
			expectError:    "LINE API returned status 400",
		},
		{
			name:           "handles 401 Unauthorized",
			statusCode:     http.StatusUnauthorized,
			serverResponse: `{"message": "Invalid access token"}`,
			expectError:    "LINE API returned status 401",
		},
		{
			name:           "handles 500 Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			serverResponse: `{"message": "Internal error"}`,
			expectError:    "LINE API returned status 500",
		},
		{
			name:           "handles 429 Too Many Requests",
			statusCode:     http.StatusTooManyRequests,
			serverResponse: `{"message": "Rate limit exceeded"}`,
			expectError:    "LINE API returned status 429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			messages := []string{"Test message"}
			err := sendLineMessage(http.DefaultClient, server.URL, "test-token", messages, "U123456")

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestSendLineMessage_NetworkError(t *testing.T) {
	// Use invalid URL to trigger network error
	messages := []string{"Test message"}
	err := sendLineMessage(http.DefaultClient, "http://invalid-host-that-does-not-exist:99999", "test-token", messages, "U123456")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestSendLineMessage_MessageLengthValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sentMessages": []}`))
	}))
	defer server.Close()

	tests := []struct {
		name        string
		messages    []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "accepts message at exactly 5000 characters",
			messages:    []string{string(make([]byte, MaxMessageLength))},
			expectError: false,
		},
		{
			name:        "rejects message exceeding 5000 characters",
			messages:    []string{string(make([]byte, MaxMessageLength+1))},
			expectError: true,
			errorMsg:    "message 1 exceeds LINE's 5000 character limit",
		},
		{
			name:        "rejects second message exceeding limit",
			messages:    []string{"short", string(make([]byte, MaxMessageLength+100))},
			expectError: true,
			errorMsg:    "message 2 exceeds LINE's 5000 character limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sendLineMessage(http.DefaultClient, server.URL, "test-token", tt.messages, "U123456")

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
