package room

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jkmoona/go-chat/server/internal/ws"
)

type Handler struct {
	svc Service
	hub *ws.Hub
}

func NewHandler(svc Service, hub *ws.Hub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("userId")
	creatorID, _ := strconv.ParseInt(userIDStr.(string), 10, 64)

	room, err := h.svc.CreateRoom(c.Request.Context(), &req, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room"})
		return
	}

	h.hub.CreateRoom(&ws.Room{
		ID:        room.ID,
		Name:      room.Name,
		Clients:   make(map[string]*ws.Client),
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
	})

	c.JSON(http.StatusCreated, RoomRes{
		ID:        room.ID,
		Name:      room.Name,
		TTL:       room.TTL,
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
		Clients:   0,
	})
}

func (h *Handler) ListRooms(c *gin.Context) {
	rooms, err := h.svc.ListActiveRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list rooms"})
		return
	}

	res := make([]RoomRes, 0, len(rooms))
	for _, r := range rooms {
		info := h.hub.GetRoomInfo(r.ID)
		res = append(res, RoomRes{
			ID:        r.ID,
			Name:      r.Name,
			TTL:       r.TTL,
			ExpiresAt: r.ExpiresAt,
			HasPIN:    r.PinHash != "",
			Clients:   info.ClientCount,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) GetRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	room, err := h.svc.GetRoom(c.Request.Context(), roomID)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get room"})
		return
	}

	info := h.hub.GetRoomInfo(room.ID)

	c.JSON(http.StatusOK, RoomRes{
		ID:        room.ID,
		Name:      room.Name,
		TTL:       room.TTL,
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
		Clients:   info.ClientCount,
	})
}

func (h *Handler) VerifyPIN(c *gin.Context) {
	roomID := c.Param("roomId")

	var req VerifyPINReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.VerifyPIN(c.Request.Context(), roomID, req.PIN); err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		if errors.Is(err, ErrInvalidPIN) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid pin"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pin verified"})
}
