package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already in use")
	ErrTokenGeneration    = errors.New("failed to generate token")
)
