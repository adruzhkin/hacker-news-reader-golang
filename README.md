# Hacker News Reader

CLI tool that fetches top Hacker News stories via the [HN Firebase API](https://github.com/HackerNews/API) and identifies the most prolific commenters per story.

For each story the output shows:
- The story title
- The top N commenters, each with their per-story comment count and total comment count across all fetched stories

### Example

```text
╭─ Cog: Containers for Machine Learning
│  ╰─ nigma1337 (1 for story - 1 total)
├─ Changing std:sort at Google's scale and beyond
│  ├─ orlp (8 for story - 8 total)
│  ├─ tsimionescu (7 for story - 8 total)
│  ├─ danlark (4 for story - 4 total)
│  ├─ jeffbee (4 for story - 4 total)
│  ╰─ samhw (4 for story - 4 total)
```

### Requirements

- Go 1.26.1+

### Build and Run

Run with default settings (30 stories, 10 commenters):
```bash
go run ./cmd/hn-reader
```

Custom number of stories and commenters:
```bash
go run ./cmd/hn-reader -story 5 -user 3
```

### Testing

```bash
go test ./...           # run all tests
go test -race ./...     # run with race detector
go test -v ./...        # verbose output
```