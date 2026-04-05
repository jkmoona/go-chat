package ws

import (
	"log/slog"
	"time"
)

type Room struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Clients   map[string]*Client `json:"clients"`
	ExpiresAt time.Time          `json:"expires_at"`
	HasPIN    bool               `json:"has_pin"`
}

type RoomInfo struct {
	Exists      bool
	HasPIN      bool
	ExpiresAt   time.Time
	ClientCount int
}

type createRoomReq struct {
	room *Room
	res  chan struct{}
}

type getRoomsReq struct {
	res chan []*Room
}

type getClientsReq struct {
	roomID string
	res    chan []*Client
}

type getRoomInfoReq struct {
	roomID string
	res    chan RoomInfo
}

type kickReq struct {
	roomID   string
	clientID string
	res      chan bool
}

type updateTTLReq struct {
	roomID    string
	expiresAt time.Time
	res       chan bool
}

type deleteRoomReq struct {
	roomID string
	res    chan bool
}

type Hub struct {
	rooms        map[string]*Room
	Register     chan *Client
	Unregister   chan *Client
	Broadcast    chan *Message
	createRoom   chan *createRoomReq
	getRooms     chan *getRoomsReq
	getClients   chan *getClientsReq
	getRoomInfo  chan *getRoomInfoReq
	expireRoom   chan string
	kick         chan *kickReq
	updateTTL    chan *updateTTLReq
	deleteRoom   chan *deleteRoomReq
	done         chan struct{}
	onRoomExpire func(string)
	idleTimers   map[string]*time.Timer
}

func NewHub(onRoomExpire func(string)) *Hub {
	return &Hub{
		rooms:        make(map[string]*Room),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan *Message, 5),
		createRoom:   make(chan *createRoomReq),
		getRooms:     make(chan *getRoomsReq),
		getClients:   make(chan *getClientsReq),
		getRoomInfo:  make(chan *getRoomInfoReq),
		expireRoom:   make(chan string, 10),
		kick:         make(chan *kickReq),
		updateTTL:    make(chan *updateTTLReq),
		deleteRoom:   make(chan *deleteRoomReq),
		done:         make(chan struct{}),
		onRoomExpire: onRoomExpire,
		idleTimers:   make(map[string]*time.Timer),
	}
}

func (h *Hub) CreateRoom(room *Room) {
	req := &createRoomReq{room: room, res: make(chan struct{})}
	h.createRoom <- req
	<-req.res
}

func (h *Hub) GetRooms() []*Room {
	req := &getRoomsReq{res: make(chan []*Room)}
	h.getRooms <- req
	return <-req.res
}

func (h *Hub) GetClients(roomID string) []*Client {
	req := &getClientsReq{roomID: roomID, res: make(chan []*Client)}
	h.getClients <- req
	return <-req.res
}

func (h *Hub) GetRoomInfo(roomID string) RoomInfo {
	req := &getRoomInfoReq{roomID: roomID, res: make(chan RoomInfo)}
	h.getRoomInfo <- req
	return <-req.res
}

func (h *Hub) Close() {
	close(h.done)
}

func (h *Hub) KickClient(roomID, clientID string) bool {
	req := &kickReq{roomID: roomID, clientID: clientID, res: make(chan bool)}
	h.kick <- req
	return <-req.res
}

func (h *Hub) UpdateRoomTTL(roomID string, expiresAt time.Time) bool {
	req := &updateTTLReq{roomID: roomID, expiresAt: expiresAt, res: make(chan bool)}
	h.updateTTL <- req
	return <-req.res
}

func (h *Hub) DeleteRoom(roomID string) bool {
	req := &deleteRoomReq{roomID: roomID, res: make(chan bool)}
	h.deleteRoom <- req
	return <-req.res
}

func (h *Hub) trySend(cl *Client, msg *Message) {
	select {
	case cl.Message <- msg:
	default:
	}
}

// drainAndSend drains any buffered messages then delivers msg, guaranteeing
// delivery for high-priority notifications (expire, delete) before channel close.
func drainAndSend(cl *Client, msg *Message) {
	for {
		select {
		case <-cl.Message:
		default:
			cl.Message <- msg
			return
		}
	}
}

func (h *Hub) broadcastPresence(room *Room) {
	seen := make(map[string]struct{})
	clients := make([]ClientInfo, 0, len(room.Clients))
	for _, cl := range room.Clients {
		if _, dup := seen[cl.ID]; dup {
			continue
		}
		seen[cl.ID] = struct{}{}
		clients = append(clients, ClientInfo{
			ID:       cl.ID,
			Username: cl.Username,
			IsGuest:  cl.IsGuest,
		})
	}

	msg := &Message{
		Type:    MessageTypePresence,
		RoomID:  room.ID,
		Clients: clients,
	}

	for _, cl := range room.Clients {
		h.trySend(cl, msg)
	}
}

func (h *Hub) doExpireRoom(roomID string) {
	room, ok := h.rooms[roomID]
	if !ok {
		return
	}

	msg := &Message{
		Content: "room has expired",
		RoomID:  roomID,
		Type:    MessageTypeExpired,
	}
	for _, cl := range room.Clients {
		drainAndSend(cl, msg)
		close(cl.Message)
	}

	delete(h.rooms, roomID)

	if t, ok := h.idleTimers[roomID]; ok {
		t.Stop()
		delete(h.idleTimers, roomID)
	}

	slog.Info("room expired", slog.String("room_id", roomID), slog.String("room_name", room.Name))

	if h.onRoomExpire != nil {
		go h.onRoomExpire(roomID)
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.done:
			slog.Info("hub shutting down, closing all connections")
			ticker.Stop()
			for _, room := range h.rooms {
				for _, cl := range room.Clients {
					close(cl.Message)
				}
			}
			return

		case <-ticker.C:
			now := time.Now()
			for id, room := range h.rooms {
				if now.After(room.ExpiresAt) {
					h.doExpireRoom(id)
					continue
				}

				remaining := int(time.Until(room.ExpiresAt).Seconds())
				if len(room.Clients) > 0 {
					msg := &Message{
						Type:      MessageTypeCountdown,
						RoomID:    room.ID,
						Remaining: remaining,
					}
					for _, cl := range room.Clients {
						h.trySend(cl, msg)
					}
				}
			}

		case roomID := <-h.expireRoom:
			h.doExpireRoom(roomID)

		case req := <-h.createRoom:
			h.rooms[req.room.ID] = req.room
			slog.Info("room created",
				slog.String("room_id", req.room.ID),
				slog.String("room_name", req.room.Name),
			)
			close(req.res)

		case req := <-h.getRooms:
			rooms := make([]*Room, 0, len(h.rooms))
			for _, r := range h.rooms {
				rooms = append(rooms, r)
			}
			req.res <- rooms

		case req := <-h.getClients:
			var clients []*Client
			if room, ok := h.rooms[req.roomID]; ok {
				clients = make([]*Client, 0, len(room.Clients))
				for _, cl := range room.Clients {
					clients = append(clients, cl)
				}
			}
			req.res <- clients

		case req := <-h.getRoomInfo:
			room, ok := h.rooms[req.roomID]
			if !ok {
				req.res <- RoomInfo{}
				continue
			}
			req.res <- RoomInfo{
				Exists:      true,
				HasPIN:      room.HasPIN,
				ExpiresAt:   room.ExpiresAt,
				ClientCount: len(room.Clients),
			}

		case cl := <-h.Register:
			if room, ok := h.rooms[cl.RoomID]; ok {
				room.Clients[cl.ConnID] = cl
				if t, ok := h.idleTimers[cl.RoomID]; ok {
					t.Stop()
					delete(h.idleTimers, cl.RoomID)
				}
				h.broadcastPresence(room)
				remaining := int(time.Until(room.ExpiresAt).Seconds())
				if remaining > 0 {
					h.trySend(cl, &Message{
						Type:      MessageTypeCountdown,
						RoomID:    room.ID,
						Remaining: remaining,
					})
				}
			}

		case cl := <-h.Unregister:
			if room, ok := h.rooms[cl.RoomID]; ok {
				if _, exists := room.Clients[cl.ConnID]; exists {
					delete(room.Clients, cl.ConnID)
					close(cl.Message)

					if len(room.Clients) > 0 {
						msg := &Message{
							Content:  cl.Username + " left the chat",
							RoomID:   cl.RoomID,
							Username: cl.Username,
							Type:     MessageTypeSystem,
						}
						for _, c := range room.Clients {
							h.trySend(c, msg)
						}
					}
					h.broadcastPresence(room)

					if len(room.Clients) == 0 {
						roomID := cl.RoomID
						h.idleTimers[roomID] = time.AfterFunc(5*time.Minute, func() {
							h.expireRoom <- roomID
						})
					}
				}
			}

		case m := <-h.Broadcast:
			if room, ok := h.rooms[m.RoomID]; ok {
				for _, cl := range room.Clients {
					if cl.ConnID != m.SenderID {
						h.trySend(cl, m)
					}
				}
			}

		case req := <-h.kick:
			room, ok := h.rooms[req.roomID]
			if !ok {
				req.res <- false
				continue
			}
			kicked := false
			var kickedUsername string
			for connID, cl := range room.Clients {
				if cl.ID == req.clientID {
					if !kicked {
						kickedUsername = cl.Username
					}
					h.trySend(cl, &Message{
						Content: "You have been removed from the room",
						RoomID:  req.roomID,
						Type:    MessageTypeKicked,
					})
					delete(room.Clients, connID)
					close(cl.Message)
					kicked = true
				}
			}
			if kicked {
				msg := &Message{
					Content:  kickedUsername + " was removed from the room",
					RoomID:   req.roomID,
					Username: kickedUsername,
					Type:     MessageTypeSystem,
				}
				for _, cl := range room.Clients {
					h.trySend(cl, msg)
				}
				h.broadcastPresence(room)
			}
			req.res <- kicked

		case req := <-h.updateTTL:
			room, ok := h.rooms[req.roomID]
			if !ok {
				req.res <- false
				continue
			}
			room.ExpiresAt = req.expiresAt
			remaining := int(time.Until(room.ExpiresAt).Seconds())
			if len(room.Clients) > 0 {
				msg := &Message{
					Type:      MessageTypeCountdown,
					RoomID:    room.ID,
					Remaining: remaining,
				}
				for _, cl := range room.Clients {
					h.trySend(cl, msg)
				}
			}
			req.res <- true

		case req := <-h.deleteRoom:
			room, ok := h.rooms[req.roomID]
			if !ok {
				req.res <- false
				continue
			}
			msg := &Message{
				Content: "room has been deleted",
				RoomID:  req.roomID,
				Type:    MessageTypeDeleted,
			}
			for _, cl := range room.Clients {
				drainAndSend(cl, msg)
				close(cl.Message)
			}
			delete(h.rooms, req.roomID)
			if t, ok := h.idleTimers[req.roomID]; ok {
				t.Stop()
				delete(h.idleTimers, req.roomID)
			}
			slog.Info("room deleted", slog.String("room_id", req.roomID), slog.String("room_name", room.Name))
			if h.onRoomExpire != nil {
				go h.onRoomExpire(req.roomID)
			}
			req.res <- true
		}
	}
}
