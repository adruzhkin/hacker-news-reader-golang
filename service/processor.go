package service

import (
	"context"
	"log"
	"sync"
)

// ProcessStory fetches a story and recursively processes its entire comment tree.
func (s *Service) ProcessStory(ctx context.Context, storyID int, wg *sync.WaitGroup) {
	defer wg.Done()

	story, err := s.FetchStory(ctx, storyID)
	if err != nil {
		log.Printf("skipping story %d: %v", storyID, err)
		return
	}

	s.MainStoryRepo.AddNew(story.ID, story)

	wg.Add(1)
	go s.ProcessAll(ctx, story.Kids, story.ID, wg)
}

// ProcessAll fans out goroutines to process each comment ID in the given slice.
func (s *Service) ProcessAll(ctx context.Context, comments []int, storyID int, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(comments) == 0 {
		return
	}

	for _, commentID := range comments {
		if ctx.Err() != nil {
			return
		}
		wg.Add(1)
		go s.Process(ctx, commentID, storyID, wg)
	}
}

// Process fetches a single comment, records its author, and recurses into child comments.
func (s *Service) Process(ctx context.Context, commentID int, storyID int, wg *sync.WaitGroup) {
	defer wg.Done()

	comment, err := s.FetchComment(ctx, commentID)
	if err != nil {
		log.Printf("skipping comment %d: %v", commentID, err)
		return
	}

	wg.Add(1)
	go s.ProcessAll(ctx, comment.Kids, storyID, wg)

	if comment.IsDeleted || comment.CreatedBy == "" {
		return
	}

	name := comment.CreatedBy
	s.MainUserRepo.IncrementCountFor(name)

	userRepo := s.MainStoryRepo.GetUsersFor(storyID)
	userRepo.IncrementCountFor(name)
}
