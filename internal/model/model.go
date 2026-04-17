package model

import "time"

type Link struct {
	ID        int64
	URL       string
	ShortURL  string
	CreatedAt time.Time
}
