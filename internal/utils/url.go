package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gianglt1/short-link/internal/domain"
)

const maxURLLen = 2048

// ValidateURL ensures the raw URL is a well-formed http/https URL within the
// length limit. It returns domain.ErrInvalidURL on failure.
func ValidateURL(rawURL string) error {
	if len(rawURL) > maxURLLen {
		return fmt.Errorf("%w: URL exceeds %d characters", domain.ErrInvalidURL, maxURLLen)
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return domain.ErrInvalidURL
	}
	return nil
}

// ExtractCode strips the base URL prefix if present, then returns the last path
// segment. It accepts a full short URL or a bare code.
func ExtractCode(input, baseURL string) string {
	input = strings.TrimPrefix(input, baseURL+"/")
	if strings.Contains(input, "/") {
		parts := strings.Split(strings.TrimRight(input, "/"), "/")
		return parts[len(parts)-1]
	}
	return input
}
