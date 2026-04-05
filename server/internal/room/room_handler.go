package room

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jkmoona/go-chat/server/internal/ws"
)

func bindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		e := ve[0]
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			return field + " is required"
		case "min":
			return field + " must be at least " + e.Param() + " characters"
		case "max":
			return field + " must be at most " + e.Param() + " characters"
		case "oneof":
			return field + " must be one of: " + strings.ReplaceAll(e.Param(), " ", ", ")
		case "numeric":
			return field + " must be numeric"
		case "len":
			return field + " must be exactly " + e.Param() + " characters"
		}
	}
	return "invalid request"
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": bindingError(err)})
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

	var requesterID int64
	if userIDStr, exists := c.Get("userId"); exists {
		requesterID, _ = strconv.ParseInt(userIDStr.(string), 10, 64)
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
			IsCreator: requesterID != 0 && r.CreatorID == requesterID,
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

	var isCreator bool
	if userIDStr, exists := c.Get("userId"); exists {
		if requesterID, err := strconv.ParseInt(userIDStr.(string), 10, 64); err == nil {
			isCreator = requesterID == room.CreatorID
		}
	}

	c.JSON(http.StatusOK, RoomRes{
		ID:        room.ID,
		Name:      room.Name,
		TTL:       room.TTL,
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
		Clients:   info.ClientCount,
		IsCreator: isCreator,
	})
}

func (h *Handler) VerifyPIN(c *gin.Context) {
	roomID := c.Param("roomId")

	var req VerifyPINReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindingError(err)})
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

func (h *Handler) requireCreator(c *gin.Context) (*Room, bool) {
	roomID := c.Param("roomId")
	userIDStr, _ := c.Get("userId")
	requesterID, _ := strconv.ParseInt(userIDStr.(string), 10, 64)

	room, err := h.svc.GetRoom(c.Request.Context(), roomID)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get room"})
		}
		return nil, false
	}

	if room.CreatorID != requesterID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the room creator can perform this action"})
		return nil, false
	}

	return room, true
}

func (h *Handler) ExtendRoom(c *gin.Context) {
	room, ok := h.requireCreator(c)
	if !ok {
		return
	}

	var req ExtendTTLReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindingError(err)})
		return
	}

	updated, err := h.svc.ExtendTTL(c.Request.Context(), room.ID, req.TTL)
	if err != nil {
		if errors.Is(err, ErrMaxLifetimeReached) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extend room"})
		return
	}

	if !h.hub.UpdateRoomTTL(room.ID, updated.ExpiresAt) {
		c.JSON(http.StatusConflict, gin.H{"error": "room has already expired"})
		return
	}

	info := h.hub.GetRoomInfo(room.ID)
	c.JSON(http.StatusOK, RoomRes{
		ID:        updated.ID,
		Name:      updated.Name,
		TTL:       updated.TTL,
		ExpiresAt: updated.ExpiresAt,
		HasPIN:    updated.PinHash != "",
		Clients:   info.ClientCount,
		IsCreator: true,
	})
}

func (h *Handler) DeleteRoom(c *gin.Context) {
	room, ok := h.requireCreator(c)
	if !ok {
		return
	}

	h.hub.DeleteRoom(room.ID)
	if err := h.svc.Deactivate(c.Request.Context(), room.ID); err != nil {
		slog.Error("failed to deactivate deleted room", slog.String("room_id", room.ID), slog.String("error", err.Error()))
	}

	c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
}

func (h *Handler) KickClient(c *gin.Context) {
	room, ok := h.requireCreator(c)
	if !ok {
		return
	}

	var req KickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindingError(err)})
		return
	}

	userIDStr, _ := c.Get("userId")
	if req.ClientID == userIDStr.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot kick yourself"})
		return
	}

	if !h.hub.KickClient(room.ID, req.ClientID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found in room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "client kicked"})
}
