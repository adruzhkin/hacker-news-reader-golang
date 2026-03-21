package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
)

const sortKey = "%22$key%22"

var (
	baseURL = "https://hacker-news.firebaseio.com/v0/topstories.json?limitToFirst=%d&orderBy=%s"
	itemURL = "https://hacker-news.firebaseio.com/v0/item/%v.json"
)

func (s *Service) FetchStoryIDs(ctx context.Context) (stories []int, err error) {
	url := fmt.Sprintf(baseURL, s.StoryLimit, sortKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return []int{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []int{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []int{}, err
	}

	err = json.Unmarshal(body, &stories)
	if err != nil {
		return []int{}, err
	}

	return stories, nil
}

func (s *Service) FetchStory(ctx context.Context, id int) (story models.Story, err error) {
	if ctx.Err() != nil {
		return models.Story{}, ctx.Err()
	}

	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	url := fmt.Sprintf(itemURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.Story{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.Story{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Story{}, err
	}

	err = json.Unmarshal(body, &story)
	if err != nil {
		return models.Story{}, err
	}

	return story, nil
}

func (s *Service) FetchComment(ctx context.Context, id int) (comment models.Comment, err error) {
	if ctx.Err() != nil {
		return models.Comment{}, ctx.Err()
	}

	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	url := fmt.Sprintf(itemURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.Comment{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.Comment{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Comment{}, err
	}

	err = json.Unmarshal(body, &comment)
	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}
