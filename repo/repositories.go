package repo

import (
	"sync"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
)

// UserRepo holds a map of users names to a number of comments they made. This
// map is used to fetch data over the http calls. User list presents in order to
// sort the same users.
type UserRepo struct {
	Mutex sync.Mutex
	Users map[string]int
	List  *models.UserList
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		Mutex: sync.Mutex{},
		Users: map[string]int{},
	}
}

func (r *UserRepo) IncrementCountFor(name string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	_, ok := r.Users[name]
	if !ok {
		r.Users[name] = 0
	}

	r.Users[name] += 1
}

// StoryRepo holds a map of stories to users who commented on that specific story.
type StoryRepo struct {
	Mutex   sync.Mutex
	Stories map[*models.Story]*UserRepo
}

func NewStoryRepo() *StoryRepo {
	return &StoryRepo{
		Mutex:   sync.Mutex{},
		Stories: map[*models.Story]*UserRepo{},
	}
}

func (r *StoryRepo) AddNew(story *models.Story) {
	userRepo := NewUserRepo()

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.Stories[story] = userRepo
}

func (r *StoryRepo) GetStories(story *models.Story) *UserRepo {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	return r.Stories[story]
}
