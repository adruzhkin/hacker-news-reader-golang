package models

type Story struct {
	ID          int    `json:"id"`
	CreatedBy   string `json:"by"`
	Descendants int    `json:"descendants"`
	Kids        []int  `json:"kids"`
	Title       string `json:"title"`
}
