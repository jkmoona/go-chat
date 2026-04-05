package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = 45 * time.Second
	writeWait  = 10 * time.Second
)

const (
	MessageTypeChat      = "chat"
	MessageTypeSystem    = "system"
	MessageTypeTyping    = "typing"
	MessageTypePresence  = "presence"
	MessageTypeCountdown = "countdown"
	MessageTypeKicked    = "kicked"
	MessageTypeExpired   = "expired"
	MessageTypeDeleted   = "deleted"
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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Message:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(msg); err != nil {
				slog.Warn("websocket write error",
					slog.String("client", c.Username),
					slog.String("room", c.RoomID),
					slog.String("error", err.Error()),
				)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
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
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

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
