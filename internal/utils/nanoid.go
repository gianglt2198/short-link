package utils

import (
	"github.com/gianglt1/short-link/internal/common"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func NewID(size int, prefix string) string {
	l := size - len(prefix)
	if l <= 0 {
		return ""
	}
	id, _ := gonanoid.Generate(common.ALPHABET, l)
	return prefix + id
}
