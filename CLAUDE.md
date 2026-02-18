# imakoko - Multi-Source News Notification App

A Go application that fetches the latest news from multiple sources (RSS, Atom, Reddit) and sends them to users through the LINE Messaging API.

## Tech Stack

- **Language:** Go 1.23.3
- **Testing:** testify v1.11.1
- **No external frameworks** — uses Go standard library (net/http, encoding/xml, encoding/json)

## Project Structure

```
main.go          - Entry point
runner.go        - Core business logic (fetch news → format → send via LINE)
config.go        - Load and validate configuration, feed source definitions
news.go          - Multi-format feed fetching and parsing (RSS, Atom, Reddit JSON)
formatter.go     - Convert news items to LINE message format (grouped by source)
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
| `FEED_SOURCES` | No | 9 built-in sources | JSON array of custom feed sources |
| `RSS_URL` | No | - | Backward compatible: single RSS feed URL |

### Feed Source Configuration

`FEED_SOURCES` accepts a JSON array of feed objects:
```json
[{"name":"HN","url":"https://hnrss.org/frontpage","type":"rss","max_items":5}]
```

Feed types: `rss`, `atom`, `reddit`

Priority: `FEED_SOURCES` > `RSS_URL` > built-in defaults

### Default Feed Sources

| Source | Type | Category |
|--------|------|----------|
| はてなブックマーク IT | RSS | Japanese tech |
| Hacker News | RSS | Global tech |
| Zenn Trending | RSS | Japanese dev community |
| Lobsters | RSS | Programming |
| Publickey | Atom | Japanese tech news |
| TechCrunch | RSS | Tech/startups |
| The Verge | Atom | Tech/general |
| r/programming | Reddit | Programming discussions |
| r/technology | Reddit | Tech news |

## Key Constants

- `MaxMessageLength = 5000` — LINE message character limit
- `maxResponseSize = 10MB` — Feed response body size limit
- `maxErrorResponseSize = 4KB` — Error response logging limit
- `batchSize = 5` — Messages per batch for LINE API

## Design Patterns

- **Interface-based design:** `MessageSender` interface for testability
- **Dependency Injection:** HTTP client injected into functions
- **Error wrapping:** `fmt.Errorf` with `%w` for error chains
- **Batch processing:** Send messages in batches of 5 per LINE API limits
- **Graceful degradation:** Partial feed failures logged and skipped, error only if all fail

## Conventions

- Comments in English
- Use testify assert/mock for tests
