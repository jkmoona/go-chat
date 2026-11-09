package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/config"
	"github.com/jkmoona/go-chat/server/internal/room"
	"github.com/jkmoona/go-chat/server/internal/user"
	"github.com/jkmoona/go-chat/server/internal/ws"
)

func init() {
	auth.Setup("test-access", "test-refresh", false)
}

// --- mock user service ---

type mockUserService struct {
	users map[string]*user.User
	nextID int64
}

func newMockUserService() *mockUserService {
	return &mockUserService{users: make(map[string]*user.User), nextID: 1}
}

func (m *mockUserService) CreateUser(_ context.Context, req *user.CreateUserReq) (*user.CreateUserRes, error) {
	if _, exists := m.users[req.Username]; exists {
		return nil, user.ErrUsernameTaken
	}
	hashed, _ := auth.HashPassword(req.Password)
	u := &user.User{ID: m.nextID, Username: req.Username, Password: hashed}
	m.users[req.Username] = u
	m.nextID++
	return &user.CreateUserRes{ID: "1", Username: u.Username}, nil
}

func (m *mockUserService) Login(_ context.Context, req *user.LoginUserReq) (*user.LoginUserRes, error) {
	u, exists := m.users[req.Username]
	if !exists {
		return nil, user.ErrInvalidCredentials
	}
	if err := auth.CheckPassword(req.Password, u.Password); err != nil {
		return nil, user.ErrInvalidCredentials
	}
	at, _ := auth.GenerateAccessToken(u.ID, u.Username, 15*time.Minute)
	rt, _ := auth.GenerateRefreshToken(u.ID, u.Username, 24*time.Hour)
	return &user.LoginUserRes{AccessToken: at, RefreshToken: rt, ID: "1", Username: u.Username}, nil
}

// --- mock room service ---

type mockRoomService struct {
	rooms map[string]*room.Room
}

func newMockRoomService() *mockRoomService {
	return &mockRoomService{rooms: make(map[string]*room.Room)}
}

func (m *mockRoomService) CreateRoom(_ context.Context, req *room.CreateRoomReq, creatorID int64) (*room.Room, error) {
	now := time.Now()
	r := &room.Room{
		ID: "abc123", Name: req.Name, CreatorID: creatorID, TTL: req.TTL,
		CreatedAt: now, ExpiresAt: now.Add(time.Duration(req.TTL) * time.Minute), IsActive: true,
	}
	m.rooms[r.ID] = r
	return r, nil
}

func (m *mockRoomService) GetRoom(_ context.Context, id string) (*room.Room, error) {
	r, ok := m.rooms[id]
	if !ok {
		return nil, room.ErrRoomNotFound
	}
	return r, nil
}

func (m *mockRoomService) ListActiveRooms(_ context.Context) ([]*room.Room, error) {
	var out []*room.Room
	for _, r := range m.rooms {
		out = append(out, r)
	}
	return out, nil
}

func (m *mockRoomService) VerifyPIN(_ context.Context, roomID, pin string) error {
	r, ok := m.rooms[roomID]
	if !ok {
		return room.ErrRoomNotFound
	}
	if r.PinHash != "" && pin != "1234" {
		return room.ErrInvalidPIN
	}
	return nil
}

func (m *mockRoomService) Deactivate(_ context.Context, id string) error {
	delete(m.rooms, id)
	return nil
}

func (m *mockRoomService) ExtendTTL(_ context.Context, roomID string, addMinutes int) (*room.Room, error) {
	r, ok := m.rooms[roomID]
	if !ok {
		return nil, room.ErrRoomNotFound
	}
	r.ExpiresAt = r.ExpiresAt.Add(time.Duration(addMinutes) * time.Minute)
	return r, nil
}

// --- test helpers ---

func setupRouter(t *testing.T) (*chi.Mux, *mockUserService, *mockRoomService) {
	t.Helper()
	userSvc := newMockUserService()
	roomSvc := newMockRoomService()

	hub := ws.NewHub(func(string) {})
	go hub.Run()
	t.Cleanup(hub.Close)

	cfg := &config.Config{ClientURL: "http://localhost:5173"}

	userHandler := user.NewHandler(userSvc)
	roomHandler := room.NewHandler(roomSvc, hub)
	wsHandler := ws.NewHandler(hub, cfg.ClientURL, func(roomID, pin string) error {
		return roomSvc.VerifyPIN(context.Background(), roomID, pin)
	})

	r := NewRouter(cfg, userHandler, roomHandler, wsHandler)
	return r, userSvc, roomSvc
}

func registerAndLogin(t *testing.T, r http.Handler) []*http.Cookie {
	t.Helper()
	body := `{"username":"testuser","password":"secret123"}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: status = %d, want %d", w.Code, http.StatusCreated)
	}

	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login: status = %d, want %d", w.Code, http.StatusOK)
	}

	return w.Result().Cookies()
}

func authedRequest(method, path string, body string, cookies []*http.Cookie) *http.Request {
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	var req *http.Request
	if reader != nil {
		req = httptest.NewRequest(method, path, reader)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	return m
}

// --- tests ---

func TestNotFoundReturnsJSON(t *testing.T) {
	r, _, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	body := decodeBody(t, w)
	if body["error"] != "not found" {
		t.Fatalf("error = %q, want %q", body["error"], "not found")
	}
}

func TestSecurityHeaders(t *testing.T) {
	r, _, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"Referrer-Policy":       "same-origin",
	}
	for key, want := range headers {
		got := w.Header().Get(key)
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if csp := w.Header().Get("Content-Security-Policy"); csp == "" {
		t.Error("Content-Security-Policy header is empty")
	}
}

func TestRegisterValidation(t *testing.T) {
	r, _, _ := setupRouter(t)

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"valid", `{"username":"validuser","password":"secret123"}`, http.StatusCreated},
		{"short username", `{"username":"ab","password":"secret123"}`, http.StatusBadRequest},
		{"short password", `{"username":"validuser2","password":"abc"}`, http.StatusBadRequest},
		{"missing username", `{"password":"secret123"}`, http.StatusBadRequest},
		{"missing password", `{"username":"validuser3"}`, http.StatusBadRequest},
		{"empty body", `{}`, http.StatusBadRequest},
		{"invalid json", `{bad`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				body := w.Body.String()
				t.Fatalf("status = %d, want %d, body = %s", w.Code, tt.status, body)
			}
		})
	}
}

func TestDuplicateRegistration(t *testing.T) {
	r, _, _ := setupRouter(t)
	body := `{"username":"dupeuser","password":"secret123"}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first register: status = %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate register: status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	r, _, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"username":"logintest","password":"correct"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"logintest","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLoginSetsCookies(t *testing.T) {
	r, _, _ := setupRouter(t)
	cookies := registerAndLogin(t, r)

	found := map[string]bool{}
	for _, c := range cookies {
		found[c.Name] = true
	}
	if !found["access_token"] || !found["refresh_token"] {
		t.Fatalf("expected access_token and refresh_token cookies, got %v", cookies)
	}
}

func TestLogoutClearsCookies(t *testing.T) {
	r, _, _ := setupRouter(t)
	cookies := registerAndLogin(t, r)

	req := authedRequest(http.MethodGet, "/logout", "", cookies)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	for _, c := range w.Result().Cookies() {
		if (c.Name == "access_token" || c.Name == "refresh_token") && c.MaxAge > 0 {
			t.Fatalf("cookie %s should be cleared, MaxAge = %d", c.Name, c.MaxAge)
		}
	}
}

func TestProtectedEndpointRequiresAuth(t *testing.T) {
	r, _, _ := setupRouter(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/ws/createRoom"},
		{http.MethodGet, "/ws/getRooms"},
		{http.MethodGet, "/ws/getClients/abc"},
		{http.MethodPatch, "/ws/room/abc"},
		{http.MethodDelete, "/ws/room/abc"},
		{http.MethodPost, "/ws/room/abc/kick"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestCreateRoomValidation(t *testing.T) {
	r, _, _ := setupRouter(t)
	cookies := registerAndLogin(t, r)

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"valid", `{"name":"room1","ttl":15}`, http.StatusCreated},
		{"missing name", `{"ttl":15}`, http.StatusBadRequest},
		{"missing ttl", `{"name":"room2"}`, http.StatusBadRequest},
		{"invalid ttl", `{"name":"room3","ttl":999}`, http.StatusBadRequest},
		{"invalid pin length", `{"name":"room4","ttl":15,"pin":"12"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := authedRequest(http.MethodPost, "/ws/createRoom", tt.body, cookies)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				body := w.Body.String()
				t.Fatalf("status = %d, want %d, body = %s", w.Code, tt.status, body)
			}
		})
	}
}

func TestGetRoomNotFound(t *testing.T) {
	r, _, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/ws/room/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetRoomPublicAccess(t *testing.T) {
	r, _, roomSvc := setupRouter(t)
	cookies := registerAndLogin(t, r)

	req := authedRequest(http.MethodPost, "/ws/createRoom", `{"name":"public-test","ttl":15}`, cookies)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	body := decodeBody(t, w)
	roomID := body["id"].(string)

	roomSvc.rooms[roomID] = roomSvc.rooms[roomID]

	req = httptest.NewRequest(http.MethodGet, "/ws/room/"+roomID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body = decodeBody(t, w)
	if body["name"] != "public-test" {
		t.Fatalf("name = %q, want %q", body["name"], "public-test")
	}
}

func TestVerifyPIN(t *testing.T) {
	r, _, roomSvc := setupRouter(t)

	roomSvc.rooms["pinroom"] = &room.Room{
		ID: "pinroom", Name: "pin room", PinHash: "something", TTL: 15,
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	req := httptest.NewRequest(http.MethodPost, "/ws/room/pinroom/verify",
		strings.NewReader(`{"pin":"1234"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("correct pin: status = %d, want %d", w.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodPost, "/ws/room/pinroom/verify",
		strings.NewReader(`{"pin":"9999"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("wrong pin: status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestDeleteRoomRequiresCreator(t *testing.T) {
	r, userSvc, roomSvc := setupRouter(t)
	cookies := registerAndLogin(t, r)

	roomSvc.rooms["other"] = &room.Room{
		ID: "other", Name: "other room", CreatorID: 999, TTL: 15,
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	_ = userSvc

	req := authedRequest(http.MethodDelete, "/ws/room/other", "", cookies)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRefreshToken(t *testing.T) {
	r, _, _ := setupRouter(t)
	cookies := registerAndLogin(t, r)

	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
		}
	}
	if refreshCookie == nil {
		t.Fatal("no refresh_token cookie")
	}

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(refreshCookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	newAccess := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "access_token" && c.Value != "" {
			newAccess = true
		}
	}
	if !newAccess {
		t.Fatal("expected new access_token cookie")
	}
}

func TestRefreshTokenMissing(t *testing.T) {
	r, _, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
