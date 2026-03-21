// Package repo provides thread-safe, in-memory repositories for stories and user comment counts.
package repo

import (
	"sort"
	"sync"

	"github.com/adruzhkin/hacker-news-reader-golang/model"
)

// UserRepo holds a map of users names to a number of comments they made. This
// map is used to fetch data over the http calls. User list presents in order to
// sort the same users.
type UserRepo struct {
	mu    sync.Mutex
	users map[string]int
	list  model.UserList
}

// NewUserRepo creates an empty UserRepo.
func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: map[string]int{},
	}
}

// IncrementCountFor atomically increments the comment count for the named user.
func (r *UserRepo) IncrementCountFor(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[name]++
}

// SortAndBuildList builds a sorted UserList from the users map.
func (r *UserRepo) SortAndBuildList() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.list = make(model.UserList, 0, len(r.users))
	for k, v := range r.users {
		r.list = append(r.list, model.User{Name: k, Count: v})
	}
	sort.Sort(r.list)
}

// GetTopUsers returns sorted users up to the given limit.
func (r *UserRepo) GetTopUsers(limit int) []model.User {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limit > len(r.list) {
		limit = len(r.list)
	}
	result := make([]model.User, limit)
	copy(result, r.list[:limit])
	return result
}

// GetCount returns the comment count for a user.
func (r *UserRepo) GetCount(name string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.users[name]
}

// StoryEntry pairs a story with its per-story user repository.
type StoryEntry struct {
	Story model.Story
	Users *UserRepo
}

// StoryRepo holds a map of stories to users who commented on that specific story.
type StoryRepo struct {
	mu      sync.Mutex
	stories map[int]*StoryEntry
}

// NewStoryRepo creates an empty StoryRepo.
func NewStoryRepo() *StoryRepo {
	return &StoryRepo{
		stories: map[int]*StoryEntry{},
	}
}

// AddNew registers a story with a fresh per-story UserRepo.
func (r *StoryRepo) AddNew(storyID int, story model.Story) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stories[storyID] = &StoryEntry{
		Story: story,
		Users: NewUserRepo(),
	}
}

// GetUsersFor returns the per-story UserRepo for the given story ID.
func (r *StoryRepo) GetUsersFor(storyID int) *UserRepo {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := r.stories[storyID]
	if entry == nil {
		return nil
	}
	return entry.Users
}

// ForEach iterates over all stories, calling fn for each one.
// The lock is held only while copying entries; fn runs unlocked.
func (r *StoryRepo) ForEach(fn func(story model.Story, users *UserRepo)) {
	r.mu.Lock()
	entries := make([]*StoryEntry, 0, len(r.stories))
	for _, entry := range r.stories {
		entries = append(entries, entry)
	}
	r.mu.Unlock()

	for _, entry := range entries {
		fn(entry.Story, entry.Users)
	}
}

// SortAllUsers calls SortAndBuildList on each per-story UserRepo.
func (r *StoryRepo) SortAllUsers() {
	r.mu.Lock()
	entries := make([]*StoryEntry, 0, len(r.stories))
	for _, entry := range r.stories {
		entries = append(entries, entry)
	}
	r.mu.Unlock()

	for _, entry := range entries {
		entry.Users.SortAndBuildList()
	}
}
