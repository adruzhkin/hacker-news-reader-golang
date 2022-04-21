package services

import (
	"sync"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
	"github.com/adruzhkin/hacker-news-reader-golang/utils"
)

func (s *Service) ProcessStory(storyID int, wg *sync.WaitGroup) {
	defer wg.Done()

	story, err := s.FetchStory(storyID)
	utils.Check(err)

	wg.Add(1)
	go s.ProcessAll(story.Kids, &story, wg)

	s.MainStoryRepo.AddNew(&story)
}

func (s *Service) ProcessAll(comments []int, story *models.Story, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(comments) == 0 {
		return
	}

	for _, commentID := range comments {
		wg.Add(1)
		go s.Process(commentID, story, wg)
	}
}

func (s *Service) Process(commentID int, story *models.Story, wg *sync.WaitGroup) {
	defer wg.Done()

	comment, err := s.FetchComment(commentID)
	utils.Check(err)
	comment.Story = story

	wg.Add(1)
	go s.ProcessAll(comment.Kids, story, wg)

	name := comment.CreatedBy
	s.MainUserRepo.IncrementCountFor(name)

	userRepo := s.MainStoryRepo.GetUsersFor(comment.Story)
	userRepo.IncrementCountFor(name)
}
