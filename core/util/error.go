package util

import (
	"errors"
	"strings"
)

var (
	ErrBlocked = errors.New("blocked")
	ErrTimeout = errors.New("timeout")
	ErrHTTP    = errors.New("http_error")
)

func ClassifyError(err error) string {
	if errors.Is(err, ErrBlocked) {
		return "blocked"
	}

	if errors.Is(err, ErrTimeout) {
		return "timeout"
	}

	if errors.Is(err, ErrHTTP) {
		return "http_error"
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "403"):
		return "blocked"
	case strings.Contains(msg, "429"):
		return "blocked"
	case strings.Contains(msg, "timeout"):
		return "timeout"
	default:
		return "http_error"
	}
}
