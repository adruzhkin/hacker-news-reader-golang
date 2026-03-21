// Package service implements the core logic: fetching HN data and processing comment trees concurrently.
package service

import (
	"net/http"

	"github.com/adruzhkin/hacker-news-reader-golang/repo"
)

// Service is used as a dependency injection manager.
type Service struct {
	StoryLimit    int
	MainStoryRepo *repo.StoryRepo
	MainUserRepo  *repo.UserRepo
	sem           chan struct{}
	client        *http.Client
}

// New creates a Service with the given configuration, repositories, and HTTP client.
func New(storyLimit int, mainStoryRepo *repo.StoryRepo, mainUserRepo *repo.UserRepo, maxConcurrency int, client *http.Client) *Service {
	return &Service{
		StoryLimit:    storyLimit,
		MainStoryRepo: mainStoryRepo,
		MainUserRepo:  mainUserRepo,
		sem:           make(chan struct{}, maxConcurrency),
		client:        client,
	}
}
