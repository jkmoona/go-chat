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

func (h *Hub) doExpireRoom(roomID string) {
	room, ok := h.rooms[roomID]
	if !ok {
		return
	}

	for _, cl := range room.Clients {
		cl.Message <- &Message{
			Content:  "room has expired",
			RoomID:   roomID,
			Username: "system",
			Type:     "system",
		}
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
				if _, exists := room.Clients[cl.ID]; !exists {
					room.Clients[cl.ID] = cl
				}
				if t, ok := h.idleTimers[cl.RoomID]; ok {
					t.Stop()
					delete(h.idleTimers, cl.RoomID)
				}
			}

		case cl := <-h.Unregister:
			if room, ok := h.rooms[cl.RoomID]; ok {
				if _, exists := room.Clients[cl.ID]; exists {
					if len(room.Clients) != 0 {
						h.Broadcast <- &Message{
							Content:  cl.Username + " left the chat",
							RoomID:   cl.RoomID,
							Username: cl.Username,
							Type:     "system",
						}
					}
					delete(room.Clients, cl.ID)
					close(cl.Message)

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
					if cl.Username != m.Username {
						cl.Message <- m
					}
				}
			}
		}
	}
}
