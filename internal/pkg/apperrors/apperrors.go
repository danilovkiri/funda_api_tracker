package apperrors

import "errors"

var (
	ErrNotFound = errors.New("requested resource was not found")
)
