package services

import "github.com/adruzhkin/hacker-news-reader-golang/repo"

// Service is used as a dependency injection manager.
type Service struct {
	StoryLimit    int
	MainStoryRepo *repo.StoryRepo
	MainUserRepo  *repo.UserRepo
}

func New(storyLimit int, mainStoryRepo *repo.StoryRepo, mainUserRepo *repo.UserRepo) *Service {
	return &Service{
		StoryLimit:    storyLimit,
		MainStoryRepo: mainStoryRepo,
		MainUserRepo:  mainUserRepo,
	}
}
