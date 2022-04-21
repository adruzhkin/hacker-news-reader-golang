# Hacker News Reader

### Service Functionality

Reader fetches the top 30 stories from the Hacker News API: https://github.com/HackerNews/API

For each of the top stories, an output contains:

- The story title

- The top 10 commenters of that story

For each commenter:

- The number of comments they made on the story

- The total number of comments they made among all the top 30 stories.

For instance, if we consider just the 3 top stories (instead of 30) and top 2 commenters (instead of 10):

| Story A            | Story B             | Story C             |
|--------------------|---------------------|---------------------|
| user-a (1 comment) | user-a (4 comments) | user-a (4 comments) |
| user-b (2 comment) | user-b (3 comments) | user-b (5 comments) |
| user-c (3 comment) | user-c (2 comments) | user-c (3 comments) |

The output to look as follows:

| Story   | 1st Top Commenter              | 2nd Top Commenter               |
|---------|--------------------------------|---------------------------------|
| Story A | user-c (3 for story - 8 total) | user-b (2 for story - 10 total) |

### Build and Run

- Dependency Golang v1.17

To run the service with default numbers of stories (30) and commenters (10):
```
go run ./cmd
```
To run with custom number of stories and commenters, i.e. 5 stories and 3 commenters:
```
go run ./cmd -story 5 -user 3
```
<br>

- Without Golang installed, you can run the binary

For Mac or Linux:
```
./hnreader
```
```
./hnreader -story 5 -user 3
```
For Windows:
```
hnreader.exe
```
```
hnreader.exe -story 5 -user 3
```