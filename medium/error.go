package medium

import "errors"

// errors
var (
	ErrNotSupported = errors.New("medium is not supported")
	ErrInvalidURL   = errors.New("invalid url")
)
