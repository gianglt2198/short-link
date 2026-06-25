package utils

import (
	"strings"
	"testing"

	"github.com/gianglt1/short-link/internal/domain"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com/path?q=1", false},
		{"valid https with port", "https://example.com:8080/path", false},
		{"empty string", "", true},
		{"no scheme", "example.com/path", true},
		{"ftp scheme", "ftp://example.com", true},
		{"no host", "https://", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"exceeds max length", "https://example.com/" + strings.Repeat("a", maxURLLen), true},
		{"exactly max length", "https://x.co/" + strings.Repeat("a", maxURLLen-len("https://x.co/")), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateURL(%q) = nil, want error", tt.url)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateURL(%q) = %v, want nil", tt.url, err)
				}
			}
		})
	}
}

func TestValidateURL_ErrorType(t *testing.T) {
	err := ValidateURL("not-a-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), domain.ErrInvalidURL.Error()) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestExtractCode(t *testing.T) {
	const base = "http://localhost:8080"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bare code", "ABC123", "ABC123"},
		{"full short URL", base + "/ABC123", "ABC123"},
		{"full URL with path prefix", base + "/some/nested/ABC123", "ABC123"},
		{"code with trailing slash", "ABC123/", "ABC123"},
		{"full URL trailing slash", base + "/ABC123/", "ABC123"},
		{"no base prefix match", "http://other.host/ABC123", "ABC123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractCode(tt.input, base)
			if got != tt.want {
				t.Errorf("ExtractCode(%q, %q) = %q, want %q", tt.input, base, got, tt.want)
			}
		})
	}
}
