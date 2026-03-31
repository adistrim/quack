package util

import "errors"

type statusCodeError interface {
	HTTPStatusCode() int
}

var (
	ErrBlocked = errors.New("blocked")
	ErrTimeout = errors.New("timeout")
	ErrHTTP    = errors.New("http_error")
)

func ClassifyError(err error) string {
	if errors.Is(err, ErrBlocked) {
		return "blocked"
	}

	var statusErr statusCodeError
	if errors.As(err, &statusErr) {
		switch statusErr.HTTPStatusCode() {
		case 403, 429:
			return "blocked"
		default:
			return "http_error"
		}
	}

	if errors.Is(err, ErrTimeout) {
		return "timeout"
	}

	if errors.Is(err, ErrHTTP) {
		return "http_error"
	}

	return "http_error"
}
