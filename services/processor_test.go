package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
	"github.com/adruzhkin/hacker-news-reader-golang/repo"
)

// setupProcessorTest creates an httptest server that routes item requests by ID,
// overrides itemURL, and returns a Service with fresh repos.
func setupProcessorTest(t *testing.T, items map[int]string) *Service {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id int
		if _, err := fmt.Sscanf(r.URL.Path, "/item/%d.json", &id); err != nil {
			w.WriteHeader(404)
			return
		}
		body, ok := items[id]
		if !ok {
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	overrideItemURL(t, server.URL)

	return New(10, repo.NewStoryRepo(), repo.NewUserRepo(), 20, &http.Client{})
}

func TestProcessStory_SingleStoryNoComments(t *testing.T) {
	svc := setupProcessorTest(t, map[int]string{
		1: `{"id":1,"by":"alice","title":"No Comments","kids":[]}`,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	users := svc.MainStoryRepo.GetUsersFor(1)
	if users == nil {
		t.Fatal("story 1 not added to repo")
	}
	// No comments, so no user counts
	if got := svc.MainUserRepo.GetCount("alice"); got != 0 {
		t.Errorf("MainUserRepo count for alice = %d, want 0 (story author not counted)", got)
	}
}

func TestProcessStory_StoryWithComments(t *testing.T) {
	svc := setupProcessorTest(t, map[int]string{
		1:   `{"id":1,"by":"alice","title":"Story","kids":[100,101]}`,
		100: `{"id":100,"by":"bob","parent":1,"kids":[]}`,
		101: `{"id":101,"by":"carol","parent":1,"kids":[]}`,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	if got := svc.MainUserRepo.GetCount("bob"); got != 1 {
		t.Errorf("MainUserRepo count for bob = %d, want 1", got)
	}
	if got := svc.MainUserRepo.GetCount("carol"); got != 1 {
		t.Errorf("MainUserRepo count for carol = %d, want 1", got)
	}

	storyUsers := svc.MainStoryRepo.GetUsersFor(1)
	if storyUsers == nil {
		t.Fatal("story 1 not in StoryRepo")
	}
	if got := storyUsers.GetCount("bob"); got != 1 {
		t.Errorf("per-story count for bob = %d, want 1", got)
	}
}

func TestProcess_RecursiveCommentTree(t *testing.T) {
	// 3-level deep: comment 100 -> 200 -> 300
	svc := setupProcessorTest(t, map[int]string{
		1:   `{"id":1,"by":"alice","title":"Deep Tree","kids":[100]}`,
		100: `{"id":100,"by":"bob","parent":1,"kids":[200]}`,
		200: `{"id":200,"by":"carol","parent":100,"kids":[300]}`,
		300: `{"id":300,"by":"dave","parent":200,"kids":[]}`,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	for _, name := range []string{"bob", "carol", "dave"} {
		if got := svc.MainUserRepo.GetCount(name); got != 1 {
			t.Errorf("MainUserRepo count for %s = %d, want 1", name, got)
		}
	}
}

func TestProcess_SkipsDeletedComment(t *testing.T) {
	svc := setupProcessorTest(t, map[int]string{
		1:   `{"id":1,"by":"alice","title":"Story","kids":[100]}`,
		100: `{"deleted":true,"id":100,"parent":1,"kids":[]}`,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	// Deleted comment's author should not be counted
	svc.MainUserRepo.SortAndBuildList()
	top := svc.MainUserRepo.GetTopUsers(10)
	if len(top) != 0 {
		t.Errorf("expected no users, got %+v", top)
	}
}

func TestProcess_SkipsEmptyByField(t *testing.T) {
	svc := setupProcessorTest(t, map[int]string{
		1:   `{"id":1,"by":"alice","title":"Story","kids":[100]}`,
		100: `{"id":100,"by":"","parent":1,"kids":[]}`,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	svc.MainUserRepo.SortAndBuildList()
	top := svc.MainUserRepo.GetTopUsers(10)
	if len(top) != 0 {
		t.Errorf("expected no users, got %+v", top)
	}
}

func TestProcessStory_FetchError_Skips(t *testing.T) {
	// No items registered — server will return 500 for story ID 1
	svc := setupProcessorTest(t, map[int]string{})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	// Story should not be in repo
	if users := svc.MainStoryRepo.GetUsersFor(1); users != nil {
		t.Error("story 1 should not be in repo after fetch failure")
	}
}

func TestProcess_FetchCommentError_Skips(t *testing.T) {
	// Story has two kids: 100 succeeds, 101 will 500
	svc := setupProcessorTest(t, map[int]string{
		1:   `{"id":1,"by":"alice","title":"Story","kids":[100,101]}`,
		100: `{"id":100,"by":"bob","parent":1,"kids":[]}`,
		// 101 not registered → 500
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 1, &wg)
	wg.Wait()

	// bob's comment should still be counted despite 101 failing
	if got := svc.MainUserRepo.GetCount("bob"); got != 1 {
		t.Errorf("MainUserRepo count for bob = %d, want 1", got)
	}
}

func TestProcessStory_ContextCancelled(t *testing.T) {
	svc := setupProcessorTest(t, map[int]string{
		1: `{"id":1,"by":"alice","title":"Story","kids":[100]}`,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(ctx, 1, &wg)
	wg.Wait()

	// Nothing should be added
	if users := svc.MainStoryRepo.GetUsersFor(1); users != nil {
		t.Error("story should not be added when context is cancelled")
	}
}

func TestProcessStory_ConcurrentMultipleStories(t *testing.T) {
	items := map[int]string{}
	// 5 stories, each with 2 comments
	for i := 1; i <= 5; i++ {
		kid1 := i*100 + 1
		kid2 := i*100 + 2
		items[i] = fmt.Sprintf(`{"id":%d,"by":"author%d","title":"Story %d","kids":[%d,%d]}`, i, i, i, kid1, kid2)
		items[kid1] = fmt.Sprintf(`{"id":%d,"by":"user_a","parent":%d,"kids":[]}`, kid1, i)
		items[kid2] = fmt.Sprintf(`{"id":%d,"by":"user_b","parent":%d,"kids":[]}`, kid2, i)
	}

	svc := setupProcessorTest(t, items)

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go svc.ProcessStory(context.Background(), i, &wg)
	}
	wg.Wait()

	// user_a commented on all 5 stories → count = 5
	if got := svc.MainUserRepo.GetCount("user_a"); got != 5 {
		t.Errorf("MainUserRepo count for user_a = %d, want 5", got)
	}
	if got := svc.MainUserRepo.GetCount("user_b"); got != 5 {
		t.Errorf("MainUserRepo count for user_b = %d, want 5", got)
	}

	// Each per-story repo should have both users with count 1
	for i := 1; i <= 5; i++ {
		storyUsers := svc.MainStoryRepo.GetUsersFor(i)
		if storyUsers == nil {
			t.Errorf("story %d not in StoryRepo", i)
			continue
		}
		if got := storyUsers.GetCount("user_a"); got != 1 {
			t.Errorf("story %d: per-story count for user_a = %d, want 1", i, got)
		}
	}
}

// verify models are correctly deserialized in integration context
func TestProcessStory_VerifiesStoryInRepo(t *testing.T) {
	svc := setupProcessorTest(t, map[int]string{
		42: `{"id":42,"by":"alice","title":"My Story","kids":[]}`,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	svc.ProcessStory(context.Background(), 42, &wg)
	wg.Wait()

	var found bool
	svc.MainStoryRepo.ForEach(func(story models.Story, users *repo.UserRepo) {
		if story.ID == 42 && story.Title == "My Story" && story.CreatedBy == "alice" {
			found = true
		}
	})
	if !found {
		t.Error("story 42 not found in StoryRepo with expected fields")
	}
}
