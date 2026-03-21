package services

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

func New(storyLimit int, mainStoryRepo *repo.StoryRepo, mainUserRepo *repo.UserRepo, maxConcurrency int, client *http.Client) *Service {
	return &Service{
		StoryLimit:    storyLimit,
		MainStoryRepo: mainStoryRepo,
		MainUserRepo:  mainUserRepo,
		sem:           make(chan struct{}, maxConcurrency),
		client:        client,
	}
}
