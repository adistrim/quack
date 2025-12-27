package util

import "strings"

func ClassifyError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "403"):
		return "blocked"
	case strings.Contains(msg, "timeout"):
		return "timeout"
	default:
		return "http_error"
	}
}
