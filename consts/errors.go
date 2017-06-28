package consts

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrNotSupported = errors.New("not supported")
	ErrUnauthorized = errors.New("unauthorized")
)
