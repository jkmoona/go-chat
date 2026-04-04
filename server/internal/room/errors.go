package room

import "errors"

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomExpired        = errors.New("room has expired")
	ErrInvalidPIN         = errors.New("invalid pin")
	ErrMaxLifetimeReached = errors.New("room cannot be extended past 24 hours")
)
