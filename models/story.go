package models

type Story struct {
	ID          int    `json:"id"`
	CreatedBy   string `json:"by"`
	Kids        []int  `json:"kids"`
	Title       string `json:"title"`
}
