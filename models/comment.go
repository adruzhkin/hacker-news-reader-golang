package models

type Comment struct {
	IsDeleted   bool   `json:"deleted"`
	ID          int    `json:"id"`
	CreatedBy   string `json:"by"`
	Descendants int    `json:"descendants"`
	Kids        []int  `json:"kids"`
	Parent      int    `json:"parent"`
}
