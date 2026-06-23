package domain

import "time"

// Link is the core entity: a mapping between a short code and an original URL.
type Link struct {
	Code        string
	OriginalURL string
	CreatedAt   time.Time
}
