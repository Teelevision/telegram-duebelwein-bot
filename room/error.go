package room

import "errors"

// errors
var (
	ErrUserUnknown         = errors.New("user unknown")
	ErrMediumUnknown       = errors.New("medium unknown")
	ErrMediumAlreadyExists = errors.New("medium already exists")
)
