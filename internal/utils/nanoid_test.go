package utils

import (
	"strings"
	"testing"

	"github.com/gianglt1/short-link/internal/common"
)

func TestNewID_Length(t *testing.T) {
	id := NewID(32, "req")
	if len(id) != 32 {
		t.Errorf("NewID(32, \"req\") len = %d, want 32", len(id))
	}
}

func TestNewID_Prefix(t *testing.T) {
	id := NewID(32, "req")
	if !strings.HasPrefix(id, "req") {
		t.Errorf("NewID(32, \"req\") = %q, want prefix \"req\"", id)
	}
}

func TestNewID_AlphabetOnly(t *testing.T) {
	id := NewID(32, "req")
	for _, ch := range id {
		if !strings.ContainsRune(common.ALPHABET, ch) {
			t.Errorf("NewID produced char %q outside alphabet", ch)
		}
	}
}

func TestNewID_PrefixTooLong(t *testing.T) {
	if got := NewID(3, "toolong"); got != "" {
		t.Errorf("NewID(3, \"toolong\") = %q, want \"\"", got)
	}
}

func TestNewID_ExactSize(t *testing.T) {
	id := NewID(5, "ab")
	if len(id) != 5 {
		t.Errorf("NewID(5, \"ab\") len = %d, want 5", len(id))
	}
}
