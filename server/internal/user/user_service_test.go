package user

import (
	"context"
	"errors"
	"testing"

	"github.com/jkmoona/go-chat/server/internal/auth"
)

type mockRepo struct {
	users    map[string]*User
	createFn func(ctx context.Context, user *User) (*User, error)
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*User)}
}

func (m *mockRepo) CreateUser(ctx context.Context, user *User) (*User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	user.ID = int64(len(m.users) + 1)
	m.users[user.Username] = user
	return user, nil
}

func (m *mockRepo) GetUser(ctx context.Context, username string) (*User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	_, ok := m.users[username]
	return ok, nil
}

func TestMain(m *testing.M) {
	_ = auth.Setup("test-access", "test-refresh", false)
	m.Run()
}

func TestCreateUser(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	res, err := svc.CreateUser(context.Background(), &CreateUserReq{
		Username: "alice",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if res.Username != "alice" {
		t.Errorf("Username = %q, want %q", res.Username, "alice")
	}
	if res.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _ = svc.CreateUser(context.Background(), &CreateUserReq{
		Username: "alice",
		Password: "password123",
	})

	_, err := svc.CreateUser(context.Background(), &CreateUserReq{
		Username: "alice",
		Password: "password456",
	})

	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestCreateUserRepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	repo := newMockRepo()
	repo.createFn = func(ctx context.Context, user *User) (*User, error) {
		return nil, repoErr
	}
	svc := NewService(repo)

	_, err := svc.CreateUser(context.Background(), &CreateUserReq{
		Username: "alice",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestLoginSuccess(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _ = svc.CreateUser(context.Background(), &CreateUserReq{
		Username: "alice",
		Password: "password123",
	})

	res, err := svc.Login(context.Background(), &LoginUserReq{
		Username: "alice",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if res.Username != "alice" {
		t.Errorf("Username = %q, want %q", res.Username, "alice")
	}
	if res.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if res.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _ = svc.CreateUser(context.Background(), &CreateUserReq{
		Username: "alice",
		Password: "password123",
	})

	_, err := svc.Login(context.Background(), &LoginUserReq{
		Username: "alice",
		Password: "wrongpass",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.Login(context.Background(), &LoginUserReq{
		Username: "nobody",
		Password: "password123",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
