package services

import (
	"context"
	"log"
	"sync"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
)

func (s *Service) ProcessStory(ctx context.Context, storyID int, wg *sync.WaitGroup) {
	defer wg.Done()

	story, err := s.FetchStory(ctx, storyID)
	if err != nil {
		log.Printf("skipping story %d: %v", storyID, err)
		return
	}

	wg.Add(1)
	go s.ProcessAll(ctx, story.Kids, &story, wg)

	s.MainStoryRepo.AddNew(&story)
}

func (s *Service) ProcessAll(ctx context.Context, comments []int, story *models.Story, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(comments) == 0 {
		return
	}

	for _, commentID := range comments {
		if ctx.Err() != nil {
			return
		}
		wg.Add(1)
		go s.Process(ctx, commentID, story, wg)
	}
}

func (s *Service) Process(ctx context.Context, commentID int, story *models.Story, wg *sync.WaitGroup) {
	defer wg.Done()

	comment, err := s.FetchComment(ctx, commentID)
	if err != nil {
		log.Printf("skipping comment %d: %v", commentID, err)
		return
	}
	comment.Story = story

	wg.Add(1)
	go s.ProcessAll(ctx, comment.Kids, story, wg)

	if comment.IsDeleted || comment.CreatedBy == "" {
		return
	}

	name := comment.CreatedBy
	s.MainUserRepo.IncrementCountFor(name)

	userRepo := s.MainStoryRepo.GetUsersFor(comment.Story)
	userRepo.IncrementCountFor(name)
}
