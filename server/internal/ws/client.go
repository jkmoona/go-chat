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
)

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
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
}

func (c *Client) writeMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

func (c *Client) readMessage(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

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
			slog.Warn("invalid message",
				slog.String("client", c.Username),
				slog.String("error", err.Error()),
			)
			continue
		}

		msg.RoomID = c.RoomID
		msg.Username = c.Username

		hub.Broadcast <- &msg
	}
}
