package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub       *Hub
	clientURL string
	verifyPIN func(roomID, pin string) error
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
	}
}

func (h *Handler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	info := h.hub.GetRoomInfo(roomID)
	if !info.Exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	if info.HasPIN {
		pin := c.Query("pin")
		if err := h.verifyPIN(roomID, pin); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid pin"})
			return
		}
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == h.clientURL
		},
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

func (h *Handler) GetClients(c *gin.Context) {
	roomID := c.Param("roomId")
	hubClients := h.hub.GetClients(roomID)
	clients := make([]ClientRes, 0, len(hubClients))

	for _, cl := range hubClients {
		clients = append(clients, ClientRes{
			ID:       cl.ID,
			Username: cl.Username,
		})
	}

	c.JSON(http.StatusOK, clients)
}
