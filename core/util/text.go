package util

import (
	"regexp"
	"strings"
)

var whitespace = regexp.MustCompile(`\s+`)

func CleanWhitespace(s string) string {
	return whitespace.ReplaceAllString(strings.TrimSpace(s), " ")
}
