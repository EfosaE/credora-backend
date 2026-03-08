package domainerr

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrMonnifyFailure    = errors.New("failed to create monnify account")

	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrUserNotFound        = errors.New("user not found")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrTokenExpired        = errors.New("token expired")
	ErrNoRowsFound         = errors.New("no rows found")
	ErrAccountNotActivated = errors.New("account not yet activated, email login disabled")

	ErrTooManyPendingJobs = errors.New("Service temporarily unavailable")
)
