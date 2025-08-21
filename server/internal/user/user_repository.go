package user

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type repository struct {
	db DBTX
}

func NewRepository(db DBTX) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	var lastInsertedId int
	query := "INSERT INTO users(username, password) VALUES ($1, $2) returning id"
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password).Scan(&lastInsertedId)
	if err != nil {
		return &User{}, err
	}

	user.ID = int64(lastInsertedId)
	return user, nil
}

func (r *repository) GetUser(ctx context.Context, username string) (*User, error) {
	u := User{}

	query := "SELECT id, username, password FROM users WHERE username = $1"
	err := r.db.QueryRowContext(ctx, query, username).Scan(&u.ID, &u.Username, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	return exists, err
}
