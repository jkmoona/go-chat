package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Handler struct {
	hub       *Hub
	clientURL string
}

type CreateRoomReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type RoomRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewHandler(h *Hub, clientURL string) *Handler {
	return &Handler{
		hub:       h,
		clientURL: clientURL,
	}
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomID, err := gonanoid.New(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate room id"})
		return
	}

	h.hub.CreateRoom(&Room{
		ID:      roomID,
		Name:    req.Name,
		Clients: make(map[string]*Client),
	})

	c.JSON(http.StatusOK, RoomRes{
		ID:   roomID,
		Name: req.Name,
	})
}

func (h *Handler) JoinRoom(c *gin.Context) {
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
	roomID := c.Param("roomId")
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
	hubRooms := h.hub.GetRooms()
	rooms := make([]RoomRes, 0, len(hubRooms))

	for _, r := range hubRooms {
		rooms = append(rooms, RoomRes{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	c.JSON(http.StatusOK, rooms)
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
