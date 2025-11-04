package dto

import "time"

type UserLinksRequest struct {
	CreatedBy string `json:"created_by"`
	UserLinksQuery
}

type UserLinksResponse struct {
	Links      []ShortlinkDTO `json:"links"`
	NextCursor string         `json:"next_cursor"`
	CreatedBy  string         `json:"created_by"`
}

type ShortlinkDTO struct {
	ShortID   string    `json:"short_id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	IsPrivate bool      `json:"is_private"`
}
