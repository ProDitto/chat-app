package ws_delivery

import (
	"context"
	"encoding/json"
	"log"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	clients        map[string]*Client // Map userID to Client
	mu             sync.RWMutex
	broadcast      chan *BroadcastPayload
	register       chan *Client
	unregister     chan *Client
	messageService usecase.MessageUseCase
	convoService   usecase.ConversationUseCase
	gameService    usecase.GameUseCase
	eventService   usecase.EventUseCase // Added EventService
}

type BroadcastPayload struct {
	Message        []byte
	ConversationID string
	SenderID       string
	RecipientID    string // For targeted messages like game invites, friend requests
	EventType      domain.EventType // What type of event this is (e.g., "new_message", "game_invite", "game_update")
	EventPayload   interface{} // The raw payload for the event service
}

func NewHub(messageService usecase.MessageUseCase, convoService usecase.ConversationUseCase, gameService usecase.GameUseCase, eventService usecase.EventUseCase) *Hub {
	return &Hub{
		broadcast:      make(chan *BroadcastPayload),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[string]*Client),
		messageService: messageService,
		convoService:   convoService,
		gameService:    gameService,
		eventService:   eventService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client connected: %s", client.UserID)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.send)
				log.Printf("Client disconnected: %s", client.UserID)
			}
			h.mu.Unlock()
		case payload := <-h.broadcast:
			var msg WebSocketMessage
			if err := json.Unmarshal(payload.Message, &msg); err != nil {
				log.Printf("error unmarshalling broadcast message: %v", err)
				continue
			}

			// Persist the event to the database for long polling / history
			// The eventService will determine who the event is for based on type and payload.
			// This assumes `payload.EventPayload` is correctly set in client.go/handler.
			if payload.EventType != "" {
				userIDsForEvent := []string{}
				switch payload.EventType {
				case domain.EventNewMessage:
					// For new messages, the event is for all participants
					participantIDs, err := h.convoService.GetParticipantIDs(context.Background(), payload.ConversationID)
					if err != nil {
						log.Printf("error getting participants for message event: %v", err)
					} else {
						userIDsForEvent = participantIDs
					}
				case domain.EventFriendRequest:
					// For friend request, event is for the recipient
					if payload.RecipientID != "" {
						userIDsForEvent = []string{payload.RecipientID}
					}
				case domain.EventGameInvite:
					// For game invite, event is for the invited player
					if payload.RecipientID != "" {
						userIDsForEvent = []string{payload.RecipientID}
					}
				// Other events (game updates, group changes, etc.) should also be handled
				default:
					// If no specific recipient logic, assume event is for the sender (or the user who triggered it)
					if payload.SenderID != "" {
						userIDsForEvent = []string{payload.SenderID}
					}
				}

				for _, userID := range userIDsForEvent {
					if err := h.eventService.CreateEvent(context.Background(), userID, payload.EventType, payload.EventPayload); err != nil {
						log.Printf("Failed to create event for user %s, type %s: %v", userID, payload.EventType, err)
					}
				}
			}

			// Then broadcast via WebSocket to active clients
			switch msg.Type {
			case "send_message":
				domainMsg, err := msg.ToDomainMessage(payload.SenderID)
				if err != nil {
					log.Printf("Error converting to domain message: %v", err)
					continue
				}

				savedMsg, err := h.messageService.SaveMessage(context.Background(), domainMsg)
				if err != nil {
					log.Printf("Error saving message: %v", err)
					continue
				}

				// Re-marshal with full sender info for clients
				wsMsgWithSender := WebSocketMessage{Type: msg.Type, Payload: nil}
				wsMsgWithSender.Payload, _ = json.Marshal(savedMsg) // Use the saved message with sender details
				finalPayload, _ := json.Marshal(wsMsgWithSender)

				participantIDs, err := h.convoService.GetParticipantIDs(context.Background(), domainMsg.ConversationID)
				if err != nil {
					log.Printf("error getting participants for convo %s: %v", domainMsg.ConversationID, err)
					continue
				}

				h.broadcastToUsersWS(participantIDs, finalPayload)

			case "game_invite", "game_update":
				// These are already handled by eventService and will be picked up by WS clients if active
				// No direct WS broadcast needed here as eventService does the work
				log.Printf("Game event type %s, hub does not directly broadcast this via WS anymore, relying on eventService", msg.Type)

			default:
				log.Printf("Unknown WebSocket message type: %s", msg.Type)
			}
		}
	}
}

// Helper to broadcast to a list of user IDs via WebSocket only
func (h *Hub) broadcastToUsersWS(userIDs []string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, userID := range userIDs {
		if client, ok := h.clients[userID]; ok {
			select {
			case client.send <- message:
				// Sent via WebSocket
			default:
				close(client.send)
				delete(h.clients, userID) // Remove problematic client
			}
		}
	}
}
