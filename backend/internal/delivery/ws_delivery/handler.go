package ws_delivery

import (
	"log"
	"net/http"
	"real-time-chat/internal/usecase"
)

type WSHandler struct {
	hub          *Hub
	tokenService usecase.TokenUseCase
}

func NewWSHandler(hub *Hub, ts usecase.TokenUseCase) *WSHandler {
	return &WSHandler{hub: hub, tokenService: ts}
}

func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusUnauthorized)
		return
	}
	claims, err := h.tokenService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: h.hub, conn: conn, send: make(chan []byte, 256), UserID: claims.UserID}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
