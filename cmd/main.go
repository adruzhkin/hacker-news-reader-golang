package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
	"github.com/adruzhkin/hacker-news-reader-golang/repo"
	"github.com/adruzhkin/hacker-news-reader-golang/services"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	mainStoryRepo = repo.NewStoryRepo()
	mainUserRepo  = repo.NewUserRepo()
	storyLimit    = flag.Int("story", 30, "how many stories to fetch")
	userLimit     = flag.Int("user", 10, "how many users to fetch for each story")
	output        = flag.String("output", "table", "type of results output")
	pool          = make(chan Job)
	service       *services.Service
	wg            sync.WaitGroup
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

	service = services.New(*storyLimit, mainStoryRepo, mainUserRepo, 20, httpClient)

	stories, err := service.FetchStoryIDs(ctx)
	if err != nil {
		log.Fatalf("failed to fetch story IDs: %v", err)
	}

	wg.Add(1)
	go allocateJobs(stories)
	go processJobs(ctx, &wg)
	wg.Wait()

	wg.Add(1)
	go processUsers(&wg)
	wg.Wait()

	switch *output {
	case "list":
		printResultsAsList()
	default:
		printResultsAsTable()
	}

	elapsed := time.Since(start)
	fmt.Printf("\nExecution took: %s\n", elapsed)
}

func allocateJobs(stories []int) {
	for _, storyID := range stories {
		pool <- Job{storyID}
	}
	close(pool)
}

func processJobs(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range pool {
		wg.Add(1)
		go service.ProcessStory(ctx, job.StoryID, wg)
	}
}

func processUsers(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, userRepo := range mainStoryRepo.Stories {
		wg.Add(1)
		go sortUsers(userRepo, wg)
	}
}

func sortUsers(userRepo *repo.UserRepo, wg *sync.WaitGroup) {
	defer wg.Done()

	userRepo.Mutex.Lock()
	defer userRepo.Mutex.Unlock()

	userList := make(models.UserList, len(userRepo.Users))
	i := 0
	for k, v := range userRepo.Users {
		userList[i] = models.User{Name: k, Count: v}
		i++
	}

	sort.Sort(userList)
	userRepo.List = &userList
}

func printResultsAsTable() {
	t := table.NewWriter()
	tTemp := table.Table{}
	tTemp.Render()

	for story, userRepo := range mainStoryRepo.Stories {
		r := table.Row{}
		r = append(r, story.Title)
		for i, user := range *userRepo.List {
			if i >= *userLimit {
				break
			}
			r = append(r, fmt.Sprintf("%s (%d for story - %d total)", user.Name, user.Count, mainUserRepo.Users[user.Name]))
		}
		t.AppendRow(r)
	}

	fmt.Println(t.Render())
}

func printResultsAsList() {
	l := list.NewWriter()
	lTemp := list.List{}
	lTemp.Render()
	l.SetStyle(list.StyleConnectedRounded)

	for story, userRepo := range mainStoryRepo.Stories {
		l.AppendItem(story.Title)
		l.Indent()
		for i, user := range *userRepo.List {
			if i >= *userLimit {
				break
			}
			l.AppendItem(fmt.Sprintf("%s (%d for story - %d total)", user.Name, user.Count, mainUserRepo.Users[user.Name]))
		}
		l.UnIndent()
	}

	fmt.Println(l.Render())
}
