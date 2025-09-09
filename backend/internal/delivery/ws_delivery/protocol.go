package ws_delivery

import (
	"encoding/json"
	"errors"
	"real-time-chat/internal/domain"
)

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SendMessagePayload struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

func (wsm *WebSocketMessage) ToDomainMessage(senderID string) (*domain.Message, error) {
	if wsm.Type != "send_message" {
		return nil, errors.New("invalid message type for ToDomainMessage")
	}
	var p SendMessagePayload
	if err := json.Unmarshal(wsm.Payload, &p); err != nil {
		return nil, err
	}
	return &domain.Message{
		ConversationID: p.ConversationID,
		SenderID:       senderID,
		Content:        p.Content,
	}, nil
}

func (wsm *WebSocketMessage) GetConversationID() string {
	if wsm.Type == "send_message" {
		var p SendMessagePayload
		if err := json.Unmarshal(wsm.Payload, &p); err != nil {
			return ""
		}
		return p.ConversationID
	}
	// Other message types might not have a conversation ID directly in their payload or need different parsing
	return ""
}
