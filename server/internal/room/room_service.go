package room

import (
	"context"
	"time"

	"github.com/jkmoona/go-chat/server/internal/auth"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type service struct {
	Repository
	timeout time.Duration
}

func NewService(repository Repository) Service {
	return &service{
		Repository: repository,
		timeout:    2 * time.Second,
	}
}

func (s *service) CreateRoom(c context.Context, req *CreateRoomReq, creatorID int64) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	id, err := gonanoid.New(6)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	room := &Room{
		ID:        id,
		Name:      req.Name,
		CreatorID: creatorID,
		TTL:       req.TTL,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(req.TTL) * time.Minute),
		IsActive:  true,
	}

	if req.PIN != "" {
		hash, err := auth.HashPassword(req.PIN)
		if err != nil {
			return nil, err
		}
		room.PinHash = hash
	}

	if err := s.Repository.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *service) GetRoom(c context.Context, id string) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	return s.Repository.GetByID(ctx, id)
}

func (s *service) ListActiveRooms(c context.Context) ([]*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	return s.Repository.ListActive(ctx)
}

func (s *service) VerifyPIN(c context.Context, roomID, pin string) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	room, err := s.Repository.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.PinHash == "" {
		return nil
	}

	if err := auth.CheckPassword(pin, room.PinHash); err != nil {
		return ErrInvalidPIN
	}
	return nil
}

func (s *service) Deactivate(c context.Context, id string) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	return s.Repository.Deactivate(ctx, id)
}
