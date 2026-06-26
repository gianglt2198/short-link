package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gianglt1/short-link/internal/common"
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

const codeLen = 6

var validCode = func() [256]bool {
	var t [256]bool
	for i := range len(common.ALPHABET) {
		t[common.ALPHABET[i]] = true
	}
	return t
}()

// IsValidCode reports whether s is a well-formed short code:
// exactly 6 characters, all from the base62 alphabet [0-9A-Za-z].
func IsValidCode(s string) bool {
	if len(s) != codeLen {
		return false
	}
	for i := range len(s) {
		if !validCode[s[i]] {
			return false
		}
	}
	return true
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
