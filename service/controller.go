package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/adruzhkin/hacker-news-reader-golang/model"
)

const sortKey = "%22$key%22"

var (
	baseURL = "https://hacker-news.firebaseio.com/v0/topstories.json?limitToFirst=%d&orderBy=%s"
	itemURL = "https://hacker-news.firebaseio.com/v0/item/%v.json"
)

// doRequest performs a single HTTP GET and returns the body bytes.
// It validates status codes and limits the response body to 1 MB.
// Returns (body, error, shouldRetry).
func (s *Service) doRequest(ctx context.Context, url string) ([]byte, error, bool) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err), false
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err), true
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 {
		return nil, fmt.Errorf("HTTP %d server error", res.StatusCode), true
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d client error", res.StatusCode), false
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("body read error: %w", err), true
	}

	return body, nil, false
}

// doWithRetry calls doRequest up to maxAttempts times with exponential backoff
// for retryable errors, respecting context cancellation.
func (s *Service) doWithRetry(ctx context.Context, url string, maxAttempts int) ([]byte, error) {
	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := range maxAttempts {
		body, err, shouldRetry := s.doRequest(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err

		if !shouldRetry || attempt == maxAttempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	return nil, lastErr
}

// FetchStoryIDs fetches the top story IDs from the HN API, limited to s.StoryLimit.
func (s *Service) FetchStoryIDs(ctx context.Context) (stories []int, err error) {
	url := fmt.Sprintf(baseURL, s.StoryLimit, sortKey)

	body, err := s.doWithRetry(ctx, url, 3)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &stories); err != nil {
		return nil, fmt.Errorf("parse error for story IDs: %w", err)
	}

	return stories, nil
}

// FetchStory fetches a single story by ID from the HN API.
func (s *Service) FetchStory(ctx context.Context, id int) (story model.Story, err error) {
	if ctx.Err() != nil {
		return model.Story{}, ctx.Err()
	}

	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	url := fmt.Sprintf(itemURL, id)

	body, err := s.doWithRetry(ctx, url, 3)
	if err != nil {
		return model.Story{}, err
	}

	if err := json.Unmarshal(body, &story); err != nil {
		return model.Story{}, fmt.Errorf("parse error for story %d: %w", id, err)
	}

	return story, nil
}

// FetchComment fetches a single comment by ID from the HN API.
func (s *Service) FetchComment(ctx context.Context, id int) (comment model.Comment, err error) {
	if ctx.Err() != nil {
		return model.Comment{}, ctx.Err()
	}

	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	url := fmt.Sprintf(itemURL, id)

	body, err := s.doWithRetry(ctx, url, 3)
	if err != nil {
		return model.Comment{}, err
	}

	if err := json.Unmarshal(body, &comment); err != nil {
		return model.Comment{}, fmt.Errorf("parse error for comment %d: %w", id, err)
	}

	return comment, nil
}
