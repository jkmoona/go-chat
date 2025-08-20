package user

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jkmoona/go-chat/server/internal/auth"
)

type service struct {
	Repository
	timeout time.Duration
}

func NewService(repository Repository) Service {
	return &service{
		repository,
		time.Duration(2) * time.Second,
	}
}

func (s *service) CreateUser(c context.Context, req *CreateUserReq) (*CreateUserRes, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &User{
		Username: req.Username,
		Password: hashedPassword,
	}

	r, err := s.Repository.CreateUser(ctx, u)

	if err != nil {
		return nil, err
	}

	res := &CreateUserRes{
		ID:       strconv.Itoa(int(r.ID)),
		Username: r.Username,
	}

	return res, nil
}

func (s *service) Login(c context.Context, req *LoginUserReq) (*LoginUserRes, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	u, err := s.Repository.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := auth.CheckPassword(req.Password, u.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	at, err := auth.GenerateAccessToken(u.ID, u.Username, 15*time.Minute)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	rt, err := auth.GenerateRefreshToken(u.ID, u.Username, 24*time.Hour)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return &LoginUserRes{
		AccessToken:  at,
		RefreshToken: rt,
		Username:     u.Username,
		ID:           strconv.Itoa(int(u.ID)),
	}, nil
}
