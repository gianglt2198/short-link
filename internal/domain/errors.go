package domain

import "errors"

var (
	// ErrInvalidURL is returned when a URL fails validation.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrLinkNotFound is returned when no link exists for a given code or URL.
	ErrLinkNotFound = errors.New("link not found")
)
