package helpers

import (
	"strings"
	"testing"

	"github.com/gianglt1/short-link/internal/common"
)

func TestBase62Encode_Length(t *testing.T) {
	for _, n := range []uint64{0, 1, 61, 62, codeSpace - 1} {
		code := Base62Encode(n)
		if len(code) != codeLen {
			t.Errorf("Base62Encode(%d) = %q, want length %d", n, code, codeLen)
		}
	}
}

func TestBase62Encode_Alphabet(t *testing.T) {
	code := Base62Encode(12345678)
	for _, ch := range code {
		if !strings.ContainsRune(common.ALPHABET, ch) {
			t.Errorf("character %q not in base62 alphabet", ch)
		}
	}
}

func TestBase62Encode_KnownValues(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "000000"},
		{1, "000001"},
		{61, "00000z"},
		{62, "000010"},
	}
	for _, tt := range tests {
		got := Base62Encode(tt.n)
		if got != tt.want {
			t.Errorf("Base62Encode(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestSnowflakeCodeGenerator_Generate(t *testing.T) {
	gen, err := NewSnowflakeCodeGenerator()
	if err != nil {
		t.Fatalf("NewSnowflakeCodeGenerator: %v", err)
	}

	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(code) != codeLen {
		t.Errorf("code length = %d, want %d", len(code), codeLen)
	}
	for _, ch := range code {
		if !strings.ContainsRune(common.ALPHABET, ch) {
			t.Errorf("character %q not in base62 alphabet", ch)
		}
	}
}

func TestSnowflakeCodeGenerator_Uniqueness(t *testing.T) {
	gen, err := NewSnowflakeCodeGenerator()
	if err != nil {
		t.Fatalf("NewSnowflakeCodeGenerator: %v", err)
	}

	seen := make(map[string]struct{}, 1000)
	for i := range 1000 {
		code, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate iteration %d: %v", i, err)
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("duplicate code %q at iteration %d", code, i)
		}
		seen[code] = struct{}{}
	}
}
