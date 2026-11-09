package room

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/httputil"
	"github.com/jkmoona/go-chat/server/internal/ws"
)

type Handler struct {
	svc Service
	hub *ws.Hub
}

func NewHandler(svc Service, hub *ws.Hub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomReq
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, httputil.BindingError(err))
		return
	}

	creatorIDInt, _ := auth.GetUserIDInt(r.Context())
	creatorIDStr, _ := auth.GetUserID(r.Context())

	room, err := h.svc.CreateRoom(r.Context(), &req, creatorIDInt)
	if err != nil {
		httputil.ErrorJSON(w, http.StatusInternalServerError, "failed to create room")
		return
	}

	h.hub.CreateRoom(&ws.Room{
		ID:        room.ID,
		Name:      room.Name,
		Clients:   make(map[string]*ws.Client),
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
		CreatorID: creatorIDStr,
	})

	httputil.WriteJSON(w, http.StatusCreated, RoomRes{
		ID:        room.ID,
		Name:      room.Name,
		TTL:       room.TTL,
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
		Clients:   0,
	})
}

func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.svc.ListActiveRooms(r.Context())
	if err != nil {
		httputil.ErrorJSON(w, http.StatusInternalServerError, "failed to list rooms")
		return
	}

	requesterID, _ := auth.GetUserIDInt(r.Context())

	res := make([]RoomRes, 0, len(rooms))
	for _, rm := range rooms {
		info := h.hub.GetRoomInfo(rm.ID)
		res = append(res, RoomRes{
			ID:        rm.ID,
			Name:      rm.Name,
			TTL:       rm.TTL,
			ExpiresAt: rm.ExpiresAt,
			HasPIN:    rm.PinHash != "",
			Clients:   info.ClientCount,
			IsCreator: requesterID != 0 && rm.CreatorID == requesterID,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	room, err := h.svc.GetRoom(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			httputil.ErrorJSON(w, http.StatusNotFound, "room not found")
			return
		}
		httputil.ErrorJSON(w, http.StatusInternalServerError, "failed to get room")
		return
	}

	info := h.hub.GetRoomInfo(room.ID)

	requesterID, _ := auth.GetUserIDInt(r.Context())
	isCreator := requesterID != 0 && requesterID == room.CreatorID

	httputil.WriteJSON(w, http.StatusOK, RoomRes{
		ID:        room.ID,
		Name:      room.Name,
		TTL:       room.TTL,
		ExpiresAt: room.ExpiresAt,
		HasPIN:    room.PinHash != "",
		Clients:   info.ClientCount,
		IsCreator: isCreator,
	})
}

func (h *Handler) VerifyPIN(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	var req VerifyPINReq
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, httputil.BindingError(err))
		return
	}

	if err := h.svc.VerifyPIN(r.Context(), roomID, req.PIN); err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			httputil.ErrorJSON(w, http.StatusNotFound, "room not found")
			return
		}
		if errors.Is(err, ErrInvalidPIN) {
			httputil.ErrorJSON(w, http.StatusForbidden, "invalid pin")
			return
		}
		httputil.ErrorJSON(w, http.StatusInternalServerError, "verification failed")
		return
	}

	httputil.MsgJSON(w, http.StatusOK, "message", "pin verified")
}

func (h *Handler) requireCreator(w http.ResponseWriter, r *http.Request) (*Room, bool) {
	roomID := chi.URLParam(r, "roomId")
	requesterID, _ := auth.GetUserIDInt(r.Context())

	room, err := h.svc.GetRoom(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			httputil.ErrorJSON(w, http.StatusNotFound, "room not found")
		} else {
			httputil.ErrorJSON(w, http.StatusInternalServerError, "failed to get room")
		}
		return nil, false
	}

	if room.CreatorID != requesterID {
		httputil.ErrorJSON(w, http.StatusForbidden, "only the room creator can perform this action")
		return nil, false
	}

	return room, true
}

func (h *Handler) ExtendRoom(w http.ResponseWriter, r *http.Request) {
	room, ok := h.requireCreator(w, r)
	if !ok {
		return
	}

	var req ExtendTTLReq
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, httputil.BindingError(err))
		return
	}

	updated, err := h.svc.ExtendTTL(r.Context(), room.ID, req.TTL)
	if err != nil {
		if errors.Is(err, ErrMaxLifetimeReached) {
			httputil.ErrorJSON(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		httputil.ErrorJSON(w, http.StatusInternalServerError, "failed to extend room")
		return
	}

	if !h.hub.UpdateRoomTTL(room.ID, updated.ExpiresAt) {
		httputil.ErrorJSON(w, http.StatusConflict, "room has already expired")
		return
	}

	info := h.hub.GetRoomInfo(room.ID)
	httputil.WriteJSON(w, http.StatusOK, RoomRes{
		ID:        updated.ID,
		Name:      updated.Name,
		TTL:       updated.TTL,
		ExpiresAt: updated.ExpiresAt,
		HasPIN:    updated.PinHash != "",
		Clients:   info.ClientCount,
		IsCreator: true,
	})
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	room, ok := h.requireCreator(w, r)
	if !ok {
		return
	}

	h.hub.DeleteRoom(room.ID)
	if err := h.svc.Deactivate(r.Context(), room.ID); err != nil {
		slog.Error("failed to deactivate deleted room", slog.String("room_id", room.ID), slog.String("error", err.Error()))
	}

	httputil.MsgJSON(w, http.StatusOK, "message", "room deleted")
}

func (h *Handler) KickClient(w http.ResponseWriter, r *http.Request) {
	room, ok := h.requireCreator(w, r)
	if !ok {
		return
	}

	var req KickReq
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.ErrorJSON(w, http.StatusBadRequest, httputil.BindingError(err))
		return
	}

	callerID, _ := auth.GetUserID(r.Context())
	if req.ClientID == callerID {
		httputil.ErrorJSON(w, http.StatusBadRequest, "cannot kick yourself")
		return
	}

	if !h.hub.KickClient(room.ID, req.ClientID) {
		httputil.ErrorJSON(w, http.StatusNotFound, "client not found in room")
		return
	}

	httputil.MsgJSON(w, http.StatusOK, "message", "client kicked")
}
