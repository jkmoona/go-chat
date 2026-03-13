package ws

import "log/slog"

type Room struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Clients map[string]*Client `json:"clients"`
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

type Hub struct {
	rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	createRoom chan *createRoomReq
	getRooms   chan *getRoomsReq
	getClients chan *getClientsReq
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
		createRoom: make(chan *createRoomReq),
		getRooms:   make(chan *getRoomsReq),
		getClients: make(chan *getClientsReq),
		done:       make(chan struct{}),
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

func (h *Hub) Close() {
	close(h.done)
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			slog.Info("hub shutting down, closing all connections")
			for _, room := range h.rooms {
				for _, cl := range room.Clients {
					close(cl.Message)
				}
			}
			return

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

		case cl := <-h.Register:
			if room, ok := h.rooms[cl.RoomID]; ok {
				if _, exists := room.Clients[cl.ID]; !exists {
					room.Clients[cl.ID] = cl
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
