package room

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jkmoona/go-chat/server/internal/auth"
)

type mockRepo struct {
	rooms map[string]*Room
}

func newMockRepo() *mockRepo {
	return &mockRepo{rooms: make(map[string]*Room)}
}

func (m *mockRepo) Create(ctx context.Context, room *Room) error {
	m.rooms[room.ID] = room
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Room, error) {
	r, ok := m.rooms[id]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return r, nil
}

func (m *mockRepo) ListActive(ctx context.Context) ([]*Room, error) {
	var rooms []*Room
	now := time.Now()
	for _, r := range m.rooms {
		if r.IsActive && r.ExpiresAt.After(now) {
			rooms = append(rooms, r)
		}
	}
	return rooms, nil
}

func (m *mockRepo) Deactivate(ctx context.Context, id string) error {
	r, ok := m.rooms[id]
	if !ok {
		return ErrRoomNotFound
	}
	r.IsActive = false
	return nil
}

func (m *mockRepo) UpdateExpiresAt(ctx context.Context, id string, expiresAt time.Time) error {
	r, ok := m.rooms[id]
	if !ok {
		return ErrRoomNotFound
	}
	r.ExpiresAt = expiresAt
	return nil
}

func TestMain(m *testing.M) {
	_ = auth.Setup("test-access", "test-refresh", false)
	m.Run()
}

func TestCreateRoom(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	room, err := svc.CreateRoom(context.Background(), &CreateRoomReq{
		Name: "Test Room",
		TTL:  60,
	}, 1)

	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}

	if room.Name != "Test Room" {
		t.Errorf("Name = %q, want %q", room.Name, "Test Room")
	}
	if room.TTL != 60 {
		t.Errorf("TTL = %d, want %d", room.TTL, 60)
	}
	if room.CreatorID != 1 {
		t.Errorf("CreatorID = %d, want %d", room.CreatorID, 1)
	}
	if len(room.ID) != 6 {
		t.Errorf("ID length = %d, want 6", len(room.ID))
	}
	if room.PinHash != "" {
		t.Error("PinHash should be empty when no PIN provided")
	}
	if room.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestCreateRoomWithPIN(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	room, err := svc.CreateRoom(context.Background(), &CreateRoomReq{
		Name: "Secret Room",
		TTL:  30,
		PIN:  "1234",
	}, 1)

	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}

	if room.PinHash == "" {
		t.Error("PinHash should not be empty when PIN provided")
	}
	if room.PinHash == "1234" {
		t.Error("PinHash should not equal plaintext PIN")
	}
}

func TestCreateRoomTTLValues(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	ttls := []int{15, 30, 60, 360, 1440}
	for _, ttl := range ttls {
		room, err := svc.CreateRoom(context.Background(), &CreateRoomReq{
			Name: "Room",
			TTL:  ttl,
		}, 1)

		if err != nil {
			t.Fatalf("CreateRoom(TTL=%d) error = %v", ttl, err)
		}

		expectedExpiry := room.CreatedAt.Add(time.Duration(ttl) * time.Minute)
		diff := room.ExpiresAt.Sub(expectedExpiry)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("TTL=%d: ExpiresAt off by %v", ttl, diff)
		}
	}
}

func TestGetRoom(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	created, _ := svc.CreateRoom(context.Background(), &CreateRoomReq{
		Name: "Room",
		TTL:  60,
	}, 1)

	room, err := svc.GetRoom(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetRoom() error = %v", err)
	}
	if room.ID != created.ID {
		t.Errorf("ID = %q, want %q", room.ID, created.ID)
	}
}

func TestGetRoomNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.GetRoom(context.Background(), "nope")
	if !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestVerifyPINCorrect(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	room, _ := svc.CreateRoom(context.Background(), &CreateRoomReq{
		Name: "Secret",
		TTL:  60,
		PIN:  "9876",
	}, 1)

	err := svc.VerifyPIN(context.Background(), room.ID, "9876")
	if err != nil {
		t.Errorf("VerifyPIN() with correct PIN: error = %v", err)
	}
}

func TestVerifyPINWrong(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	room, _ := svc.CreateRoom(context.Background(), &CreateRoomReq{
		Name: "Secret",
		TTL:  60,
		PIN:  "9876",
	}, 1)

	err := svc.VerifyPIN(context.Background(), room.ID, "0000")
	if !errors.Is(err, ErrInvalidPIN) {
		t.Errorf("expected ErrInvalidPIN, got %v", err)
	}
}

func TestVerifyPINNoPIN(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	room, _ := svc.CreateRoom(context.Background(), &CreateRoomReq{
		Name: "Open",
		TTL:  60,
	}, 1)

	err := svc.VerifyPIN(context.Background(), room.ID, "")
	if err != nil {
		t.Errorf("VerifyPIN() on room without PIN: error = %v", err)
	}
}

func TestVerifyPINRoomNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	err := svc.VerifyPIN(context.Background(), "nope", "1234")
	if !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestListActiveRooms(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _ = svc.CreateRoom(context.Background(), &CreateRoomReq{Name: "A", TTL: 60}, 1)
	_, _ = svc.CreateRoom(context.Background(), &CreateRoomReq{Name: "B", TTL: 60}, 1)

	rooms, err := svc.ListActiveRooms(context.Background())
	if err != nil {
		t.Fatalf("ListActiveRooms() error = %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("got %d rooms, want 2", len(rooms))
	}
}

func TestDeactivate(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	room, _ := svc.CreateRoom(context.Background(), &CreateRoomReq{Name: "A", TTL: 60}, 1)

	err := svc.Deactivate(context.Background(), room.ID)
	if err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}

	rooms, _ := svc.ListActiveRooms(context.Background())
	if len(rooms) != 0 {
		t.Errorf("deactivated room should not appear in active list, got %d rooms", len(rooms))
	}
}
