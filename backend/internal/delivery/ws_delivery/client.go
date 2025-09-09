package ws_delivery

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"real-time-chat/internal/domain" // Import domain to use EventType
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait  = 60 * time.Second
	// Ping period is 30 seconds as per requirement (must be less than pongWait)
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for development
	},
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	UserID string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			payload := &BroadcastPayload{
				Message:      message,
				SenderID:     c.UserID,
				EventType:    domain.EventType(msg.Type), // Convert string type to domain.EventType
				EventPayload: msg.Payload,                // The raw JSON payload for event persistence
			}
			// Extract ConversationID if it's a message type
			if msg.Type == "send_message" {
				payload.ConversationID = msg.GetConversationID()
			}
			// Extract RecipientID if it's a game invite
			if msg.Type == "game_invite" {
				var gameInvitePayload struct {
					Player2ID string `json:"player2_id"` // Assuming payload contains this for invites
				}
				if err := json.Unmarshal(msg.Payload, &gameInvitePayload); err == nil {
					payload.RecipientID = gameInvitePayload.Player2ID
				}
			}

			c.hub.broadcast <- payload
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
