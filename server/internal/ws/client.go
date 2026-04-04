package ws

import (
	"encoding/json"
	"log/slog"

	"github.com/gorilla/websocket"
)

const (
	MessageTypeChat      = "chat"
	MessageTypeSystem    = "system"
	MessageTypeTyping    = "typing"
	MessageTypePresence  = "presence"
	MessageTypeCountdown = "countdown"
	MessageTypeKicked    = "kicked"
)

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	ConnID   string
	ID       string `json:"id"`
	RoomID   string `json:"roomId"`
	Username string `json:"username"`
	IsGuest  bool   `json:"is_guest"`
}

type ClientInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsGuest  bool   `json:"is_guest"`
}

type Message struct {
	Content   string       `json:"content"`
	RoomID    string       `json:"roomId"`
	Username  string       `json:"username"`
	Type      string       `json:"type"`
	Clients   []ClientInfo `json:"clients,omitempty"`
	Remaining int          `json:"remaining,omitempty"`
	SenderID  string       `json:"-"`
}

func (c *Client) writeMessage() {
	defer c.Conn.Close()

	for msg := range c.Message {
		if err := c.Conn.WriteJSON(msg); err != nil {
			return
		}
	}
}

const maxMessageSize = 4096

func (c *Client) readMessage(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error",
					slog.String("client", c.Username),
					slog.String("room", c.RoomID),
					slog.String("error", err.Error()),
				)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(m, &msg); err != nil {
			continue
		}

		if len(msg.Content) == 0 || len(msg.Content) > 2000 {
			continue
		}

		msg.Type = MessageTypeChat
		msg.RoomID = c.RoomID
		msg.Username = c.Username
		msg.SenderID = c.ConnID

		hub.Broadcast <- &msg
	}
}
