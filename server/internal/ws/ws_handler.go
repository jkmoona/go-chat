package ws

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Handler struct {
	hub *Hub
}
type CreateRoomReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoomRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room name must be between 1 and 100 characters"})
		return
	}

	roomID, err := gonanoid.New(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate room id"})
		return
	}

	room := &Room{
		ID:      roomID,
		Name:    name,
		Clients: make(map[string]*Client),
	}
	h.hub.CreateRoom <- room

	c.JSON(http.StatusOK, RoomRes{
		ID:   roomID,
		Name: name,
	})
}

func (h *Handler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	h.hub.mu.RLock()
	_, exists := h.hub.Rooms[roomID]
	h.hub.mu.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clientID := c.Query("userId")
	username := c.Query("username")

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		ID:       clientID,
		RoomID:   roomID,
		Username: username,
	}

	m := &Message{
		Content:  username + " has joined the room",
		RoomID:   roomID,
		Username: username,
		Type:     "system",
	}

	h.hub.Register <- cl
	h.hub.Broadcast <- m

	go cl.writeMessage()
	cl.readMessage(h.hub)
}

func (h *Handler) GetRooms(c *gin.Context) {
	rooms := make([]RoomRes, 0)

	h.hub.mu.RLock()
	for _, r := range h.hub.Rooms {
		rooms = append(rooms, RoomRes{
			ID:   r.ID,
			Name: r.Name,
		})
	}
	h.hub.mu.RUnlock()

	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) GetClients(c *gin.Context) {
	var clients []ClientRes
	roomId := c.Param("roomId")

	h.hub.mu.RLock()
	room, ok := h.hub.Rooms[roomId]
	if !ok {
		h.hub.mu.RUnlock()
		clients = make([]ClientRes, 0)
		c.JSON(http.StatusOK, clients)
		return
	}

	for _, cl := range room.Clients {
		clients = append(clients, ClientRes{
			ID:       cl.ID,
			Username: cl.Username,
		})
	}
	h.hub.mu.RUnlock()

	c.JSON(http.StatusOK, clients)
}

func checkOrigin(r *http.Request) bool {
	clientURL := os.Getenv("CLIENT_URL")
	if clientURL == "" {
		clientURL = "http://localhost:5173"
	}
	return r.Header.Get("Origin") == clientURL
}
