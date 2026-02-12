# imakoko - Latest News Notification App

A Go application that fetches the latest news from Hacker News via RSS and sends them to users through the LINE Messaging API.

## Tech Stack

- **Language:** Go 1.23.3
- **Testing:** testify v1.11.1
- **No external frameworks** — uses Go standard library (net/http, encoding/xml, encoding/json)

## Project Structure

```
main.go          - Entry point
runner.go        - Core business logic (fetch news → format → send via LINE)
config.go        - Load and validate configuration from environment variables
news.go          - RSS feed fetching and parsing
formatter.go     - Convert news items to LINE message format
line.go          - LINE Messaging API client with batch sending
*_test.go        - Tests for each module
```

## Build / Run / Test

```bash
# Build
go build -o imakoko

# Run (requires environment variables)
./imakoko

# Test
go test ./...

# Test with coverage
go test -cover ./...
```

## Environment Variables

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `LINE_ACCESS_TOKEN` | Yes | - | LINE channel access token |
| `TARGET_USER_ID` | Yes | - | Target user ID for message delivery |
| `LINE_API_URL` | No | `https://api.line.me/v2/bot/message/push` | LINE API endpoint |
| `RSS_URL` | No | `https://hnrss.org/frontpage` | RSS feed URL |

## Key Constants

- `MaxMessageLength = 5000` — LINE message character limit
- `maxResponseSize = 10MB` — RSS response body size limit
- `maxErrorResponseSize = 4KB` — Error response logging limit
- `batchSize = 5` — Messages per batch for LINE API

## Design Patterns

- **Interface-based design:** `MessageSender` interface for testability
- **Dependency Injection:** HTTP client injected into functions
- **Error wrapping:** `fmt.Errorf` with `%w` for error chains
- **Batch processing:** Send messages in batches of 5 per LINE API limits

## Conventions

- Comments in English
- Use testify assert/mock for tests
