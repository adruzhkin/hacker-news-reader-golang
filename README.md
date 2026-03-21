# Hacker News Reader

CLI tool that fetches top Hacker News stories via the [HN Firebase API](https://github.com/HackerNews/API) and identifies the most prolific commenters per story.

For each story the output shows:
- The story title
- The top N commenters, each with their per-story comment count and total comment count across all fetched stories

### Example

Considering the top 3 stories and top 2 commenters:

| Story   | 1st Top Commenter              | 2nd Top Commenter               |
|---------|--------------------------------|---------------------------------|
| Story A | user-c (3 for story - 8 total) | user-b (2 for story - 10 total) |
| Story B | user-a (4 for story - 9 total) | user-b (3 for story - 10 total) |
| Story C | user-b (5 for story - 10 total)| user-a (4 for story - 9 total)  |

### Requirements

- Go 1.26.1+

### Build and Run

Run with default settings (30 stories, 10 commenters, table output):
```bash
go run ./cmd
```

Custom number of stories and commenters:
```bash
go run ./cmd -story 5 -user 3
```

List output instead of table:
```bash
go run ./cmd -story 5 -user 3 -output list
```

List output example:
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

### Testing

```bash
go test ./...           # run all tests
go test -race ./...     # run with race detector
go test -v ./...        # verbose output
```