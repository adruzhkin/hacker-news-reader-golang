package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
	"github.com/adruzhkin/hacker-news-reader-golang/repo"
	"github.com/adruzhkin/hacker-news-reader-golang/services"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	storyLimit = flag.Int("story", 30, "how many stories to fetch")
	userLimit  = flag.Int("user", 10, "how many users to fetch for each story")
	pool       = make(chan Job)
)

type Job struct {
	StoryID int
}

func main() {
	flag.Parse()

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConnsPerHost: 20,
		},
	}

	mainStoryRepo := repo.NewStoryRepo()
	mainUserRepo := repo.NewUserRepo()
	service := services.New(*storyLimit, mainStoryRepo, mainUserRepo, 20, httpClient)

	stories, err := service.FetchStoryIDs(ctx)
	if err != nil {
		log.Fatalf("failed to fetch story IDs: %v", err)
	}

	var wg sync.WaitGroup

	go allocateJobs(stories)
	for job := range pool {
		wg.Add(1)
		go service.ProcessStory(ctx, job.StoryID, &wg)
	}
	wg.Wait()

	mainStoryRepo.SortAllUsers()

	printResultsAsList(mainStoryRepo, mainUserRepo, *userLimit)

	elapsed := time.Since(start)
	fmt.Printf("\nExecution took: %s\n", elapsed)
}

func allocateJobs(stories []int) {
	for _, storyID := range stories {
		pool <- Job{storyID}
	}
	close(pool)
}

func printResultsAsList(storyRepo *repo.StoryRepo, userRepo *repo.UserRepo, limit int) {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	storyRepo.ForEach(func(story models.Story, users *repo.UserRepo) {
		l.AppendItem(text.FgCyan.Sprint(story.Title))
		l.Indent()
		for _, user := range users.GetTopUsers(limit) {
			l.AppendItem(text.FgGreen.Sprintf("%s (%d for story - %d total)", user.Name, user.Count, userRepo.GetCount(user.Name)))
		}
		l.UnIndent()
	})

	fmt.Println(l.Render())
}
