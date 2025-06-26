package domainerrors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrMonnifyFailure  = errors.New("failed to create monnify account")
)
