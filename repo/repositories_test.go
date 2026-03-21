package repo

import (
	"sync"
	"testing"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
)

// --- UserRepo tests ---

func TestUserRepo_IncrementAndGetCount(t *testing.T) {
	r := NewUserRepo()
	r.IncrementCountFor("alice")
	r.IncrementCountFor("alice")
	r.IncrementCountFor("alice")

	if got := r.GetCount("alice"); got != 3 {
		t.Errorf("GetCount(alice) = %d, want 3", got)
	}
	if got := r.GetCount("bob"); got != 0 {
		t.Errorf("GetCount(bob) = %d, want 0", got)
	}
}

func TestUserRepo_SortAndGetTopUsers(t *testing.T) {
	r := NewUserRepo()
	for range 5 {
		r.IncrementCountFor("alice")
	}
	for range 3 {
		r.IncrementCountFor("bob")
	}
	for range 3 {
		r.IncrementCountFor("carol")
	}

	r.SortAndBuildList()
	top := r.GetTopUsers(3)

	if len(top) != 3 {
		t.Fatalf("GetTopUsers(3) returned %d users, want 3", len(top))
	}
	if top[0].Name != "alice" || top[0].Count != 5 {
		t.Errorf("top[0] = %+v, want {alice, 5}", top[0])
	}
	// bob and carol tied at 3, alphabetical order
	if top[1].Name != "bob" || top[2].Name != "carol" {
		t.Errorf("tiebreak: got %s, %s; want bob, carol", top[1].Name, top[2].Name)
	}
}

func TestUserRepo_GetTopUsers_LimitExceedsSize(t *testing.T) {
	r := NewUserRepo()
	r.IncrementCountFor("alice")
	r.SortAndBuildList()

	top := r.GetTopUsers(100)
	if len(top) != 1 {
		t.Errorf("GetTopUsers(100) returned %d users, want 1", len(top))
	}
}

func TestUserRepo_GetTopUsers_LimitZero(t *testing.T) {
	r := NewUserRepo()
	r.IncrementCountFor("alice")
	r.SortAndBuildList()

	top := r.GetTopUsers(0)
	if len(top) != 0 {
		t.Errorf("GetTopUsers(0) returned %d users, want 0", len(top))
	}
}

func TestUserRepo_ConcurrentIncrements(t *testing.T) {
	r := NewUserRepo()
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				r.IncrementCountFor("alice")
			}
		}()
	}
	wg.Wait()

	if got := r.GetCount("alice"); got != 10000 {
		t.Errorf("GetCount(alice) = %d, want 10000", got)
	}
}

// --- StoryRepo tests ---

func TestStoryRepo_AddNewAndGetUsersFor(t *testing.T) {
	r := NewStoryRepo()
	story := models.Story{ID: 1, Title: "Test Story"}
	r.AddNew(1, story)

	users := r.GetUsersFor(1)
	if users == nil {
		t.Fatal("GetUsersFor(1) returned nil, want non-nil UserRepo")
	}

	// Verify the UserRepo is functional
	users.IncrementCountFor("alice")
	if got := users.GetCount("alice"); got != 1 {
		t.Errorf("user count = %d, want 1", got)
	}
}

func TestStoryRepo_GetUsersFor_NonExistent(t *testing.T) {
	r := NewStoryRepo()
	if got := r.GetUsersFor(999); got != nil {
		t.Errorf("GetUsersFor(999) = %v, want nil", got)
	}
}

func TestStoryRepo_ForEach(t *testing.T) {
	r := NewStoryRepo()
	r.AddNew(1, models.Story{ID: 1, Title: "Story 1"})
	r.AddNew(2, models.Story{ID: 2, Title: "Story 2"})
	r.AddNew(3, models.Story{ID: 3, Title: "Story 3"})

	visited := map[int]bool{}
	r.ForEach(func(story models.Story, users *UserRepo) {
		visited[story.ID] = true
	})

	if len(visited) != 3 {
		t.Errorf("ForEach visited %d stories, want 3", len(visited))
	}
	for _, id := range []int{1, 2, 3} {
		if !visited[id] {
			t.Errorf("ForEach did not visit story %d", id)
		}
	}
}

func TestStoryRepo_SortAllUsers(t *testing.T) {
	r := NewStoryRepo()
	r.AddNew(1, models.Story{ID: 1})
	r.AddNew(2, models.Story{ID: 2})

	users1 := r.GetUsersFor(1)
	for range 3 {
		users1.IncrementCountFor("bob")
	}
	users1.IncrementCountFor("alice")

	users2 := r.GetUsersFor(2)
	for range 2 {
		users2.IncrementCountFor("carol")
	}

	r.SortAllUsers()

	top1 := r.GetUsersFor(1).GetTopUsers(2)
	if top1[0].Name != "bob" || top1[0].Count != 3 {
		t.Errorf("story 1 top user = %+v, want {bob, 3}", top1[0])
	}

	top2 := r.GetUsersFor(2).GetTopUsers(1)
	if top2[0].Name != "carol" || top2[0].Count != 2 {
		t.Errorf("story 2 top user = %+v, want {carol, 2}", top2[0])
	}
}

func TestStoryRepo_ConcurrentAddAndRead(t *testing.T) {
	r := NewStoryRepo()
	var wg sync.WaitGroup

	// Concurrent writers
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.AddNew(i, models.Story{ID: i, Title: "Story"})
		}()
	}

	// Concurrent readers
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.GetUsersFor(i) // may return nil, that's fine
		}()
	}

	wg.Wait()
}
