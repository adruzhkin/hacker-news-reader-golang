package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/adruzhkin/hacker-news-reader-golang/models"
)

const sortKey = "%22$key%22"

var (
	baseURL = "https://hacker-news.firebaseio.com/v0/topstories.json?limitToFirst=%d&orderBy=%s"
	itemURL = "https://hacker-news.firebaseio.com/v0/item/%v.json"
)

func (s *Service) FetchStoryIDs() (stories []int, err error) {
	url := fmt.Sprintf(baseURL, s.StoryLimit, sortKey)

	res, err := http.Get(url)
	if err != nil {
		return []int{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []int{}, err
	}

	err = json.Unmarshal(body, &stories)
	if err != nil {
		return []int{}, err
	}

	return stories, nil
}

func (s *Service) FetchStory(id int) (story models.Story, err error) {
	url := fmt.Sprintf(itemURL, id)

	res, err := http.Get(url)
	if err != nil {
		return models.Story{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return models.Story{}, err
	}

	err = json.Unmarshal(body, &story)
	if err != nil {
		return models.Story{}, err
	}

	return story, nil
}

func (s *Service) FetchComment(id int) (comment models.Comment, err error) {
	url := fmt.Sprintf(itemURL, id)

	res, err := http.Get(url)
	if err != nil {
		return models.Comment{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return models.Comment{}, err
	}

	err = json.Unmarshal(body, &comment)
	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}
