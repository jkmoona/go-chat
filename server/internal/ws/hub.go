package ws

import "sync"

type Room struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Clients map[string]*Client `json:"clients"`
}

type Hub struct {
	Rooms      map[string]*Room
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	CreateRoom chan *Room
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
		CreateRoom: make(chan *Room),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case room := <-h.CreateRoom:
			h.mu.Lock()
			h.Rooms[room.ID] = room
			h.mu.Unlock()
		case cl := <-h.Register:
			h.mu.Lock()
			if r, ok := h.Rooms[cl.RoomID]; ok {
				if _, exists := r.Clients[cl.ID]; !exists {
					r.Clients[cl.ID] = cl
				}
			}
			h.mu.Unlock()
		case cl := <-h.Unregister:
			var leaveMsg *Message
			h.mu.Lock()
			if r, ok := h.Rooms[cl.RoomID]; ok {
				if _, exists := r.Clients[cl.ID]; exists {
					if len(r.Clients) != 0 {
						leaveMsg = &Message{
							Content:  cl.Username + " left the chat",
							RoomID:   cl.RoomID,
							Username: cl.Username,
							Type:     "system",
						}
					}
					delete(r.Clients, cl.ID)
					close(cl.Message)
					if len(r.Clients) == 0 {
						delete(h.Rooms, cl.RoomID)
					}
				}
			}
			h.mu.Unlock()
			if leaveMsg != nil {
				h.Broadcast <- leaveMsg
			}

		case m := <-h.Broadcast:
			h.mu.RLock()
			if r, ok := h.Rooms[m.RoomID]; ok {
				for _, cl := range r.Clients {
					if cl.Username != m.Username {
						select {
						case cl.Message <- m:
						default:
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}
