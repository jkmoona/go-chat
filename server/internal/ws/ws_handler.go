package ws

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/httputil"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Handler struct {
	hub       *Hub
	clientURL string
	verifyPIN func(roomID, pin string) error
	upgrader  websocket.Upgrader
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewHandler(h *Hub, clientURL string, verifyPIN func(roomID, pin string) error) *Handler {
	return &Handler{
		hub:       h,
		clientURL: clientURL,
		verifyPIN: verifyPIN,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == clientURL
			},
		},
	}
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	info := h.hub.GetRoomInfo(roomID)
	if !info.Exists {
		httputil.ErrorJSON(w, http.StatusNotFound, "room not found")
		return
	}

	clientID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.ErrorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if info.HasPIN && clientID != info.CreatorID {
		pin := r.URL.Query().Get("pin")
		if err := h.verifyPIN(roomID, pin); err != nil {
			httputil.ErrorJSON(w, http.StatusForbidden, "invalid pin")
			return
		}
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, "connection upgrade failed")
		return
	}
	username, ok := auth.GetUsername(r.Context())
	if !ok {
		conn.Close()
		return
	}

	connID, err := gonanoid.New(12)
	if err != nil {
		conn.Close()
		return
	}

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		ConnID:   connID,
		ID:       clientID,
		RoomID:   roomID,
		Username: username,
	}

	m := &Message{
		Content:  cl.Username + " has joined the room",
		RoomID:   roomID,
		Username: cl.Username,
		Type:     MessageTypeSystem,
		SenderID: cl.ConnID,
	}

	h.hub.Register <- cl
	h.hub.Broadcast <- m

	go cl.writeMessage()
	cl.readMessage(h.hub)
}

func sanitizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsPrint(r) && !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func (h *Handler) GuestJoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	name := sanitizeName(r.URL.Query().Get("name"))

	if len(name) == 0 || len(name) > 30 {
		httputil.ErrorJSON(w, http.StatusBadRequest, "name must be between 1 and 30 characters")
		return
	}

	info := h.hub.GetRoomInfo(roomID)
	if !info.Exists {
		httputil.ErrorJSON(w, http.StatusNotFound, "room not found")
		return
	}

	if info.HasPIN {
		pin := r.URL.Query().Get("pin")
		if err := h.verifyPIN(roomID, pin); err != nil {
			httputil.ErrorJSON(w, http.StatusForbidden, "invalid pin")
			return
		}
	}

	guestID, err := gonanoid.New(8)
	if err != nil {
		httputil.ErrorJSON(w, http.StatusInternalServerError, "failed to generate guest id")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, "connection upgrade failed")
		return
	}

	connID, err := gonanoid.New(12)
	if err != nil {
		conn.Close()
		return
	}

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		ConnID:   connID,
		ID:       "guest_" + guestID,
		RoomID:   roomID,
		Username: name,
		IsGuest:  true,
	}

	m := &Message{
		Content:  name + " has joined the room",
		RoomID:   roomID,
		Username: name,
		Type:     MessageTypeSystem,
		SenderID: cl.ConnID,
	}

	h.hub.Register <- cl
	h.hub.Broadcast <- m

	go cl.writeMessage()
	cl.readMessage(h.hub)
}

func (h *Handler) GetClients(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	hubClients := h.hub.GetClients(roomID)
	clients := make([]ClientRes, 0, len(hubClients))
	for _, cl := range hubClients {
		clients = append(clients, ClientRes{
			ID:       cl.ID,
			Username: cl.Username,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, clients)
}
