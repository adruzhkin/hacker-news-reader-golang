package model

// Comment represents a Hacker News comment item.
type Comment struct {
	IsDeleted   bool   `json:"deleted"`
	ID          int    `json:"id"`
	CreatedBy   string `json:"by"`
	Kids        []int  `json:"kids"`
	Parent      int    `json:"parent"`
}
