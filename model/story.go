// Package model defines the domain types for Hacker News stories, comments, and users.
package model

// Story represents a Hacker News story item.
type Story struct {
	ID          int    `json:"id"`
	CreatedBy   string `json:"by"`
	Kids        []int  `json:"kids"`
	Title       string `json:"title"`
}
