package user

import "context"

type User struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateUserRes struct {
	ID       string `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
}

type LoginUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUserRes struct {
	AccessToken  string
	RefreshToken string
	ID           string `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
}

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, username string) (*User, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}

type Service interface {
	CreateUser(c context.Context, req *CreateUserReq) (*CreateUserRes, error)
	Login(c context.Context, req *LoginUserReq) (*LoginUserRes, error)
}
