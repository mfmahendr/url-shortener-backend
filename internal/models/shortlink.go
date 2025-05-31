package models

import "time"

type Shortlink struct {
	ShortID   string    `firestore:"short_id"`
	URL       string    `firestore:"url"`
	CreatedAt time.Time `firestore:"created_at"`
	CreatedBy string    `firestore:"created_by"`
}