package room

import (
	"context"
	"time"
)

type Room struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatorID  int64     `json:"creator_id"`
	PinHash    string    `json:"-"`
	TTL        int       `json:"ttl_minutes"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsActive   bool      `json:"is_active"`
}

type CreateRoomReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
	TTL  int    `json:"ttl" binding:"required,oneof=15 30 60 360 1440"`
	PIN  string `json:"pin,omitempty" binding:"omitempty,numeric,len=4"`
}

type RoomRes struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TTL       int       `json:"ttl"`
	ExpiresAt time.Time `json:"expires_at"`
	HasPIN    bool      `json:"has_pin"`
	Clients   int       `json:"clients"`
}

type VerifyPINReq struct {
	PIN string `json:"pin" binding:"required"`
}

type Repository interface {
	Create(ctx context.Context, room *Room) error
	GetByID(ctx context.Context, id string) (*Room, error)
	ListActive(ctx context.Context) ([]*Room, error)
	Deactivate(ctx context.Context, id string) error
}

type Service interface {
	CreateRoom(ctx context.Context, req *CreateRoomReq, creatorID int64) (*Room, error)
	GetRoom(ctx context.Context, id string) (*Room, error)
	ListActiveRooms(ctx context.Context) ([]*Room, error)
	VerifyPIN(ctx context.Context, roomID, pin string) error
	Deactivate(ctx context.Context, id string) error
}
