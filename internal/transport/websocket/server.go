package ws

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat/internal/auth"
	"chat/internal/chat"
	"chat/internal/dialogs"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	hub        *chat.Hub
	repo       *chat.Repository
	dialogRepo *dialogs.Repository
}

func NewServer(hub *chat.Hub, repo *chat.Repository, dialogRepo *dialogs.Repository) *Server {
	return &Server{
		hub:        hub,
		repo:       repo,
		dialogRepo: dialogRepo,
	}
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	token = strings.TrimPrefix(token, "Bearer ")
	dialogIDStr := r.URL.Query().Get("dialog_id")

	if token == "" || dialogIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dialogID, err := strconv.Atoi(dialogIDStr)
	if err != nil {
		http.Error(w, "invalid dialog_id", http.StatusBadRequest)
		return
	}

	claims, err := auth.ParseToken(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	member, err := s.dialogRepo.IsMember(dialogID, claims.Username)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if !member {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &chat.Client{
		ID:       claims.Username,
		DialogID: dialogID,
		Conn:     conn,
		Send:     make(chan chat.Message, 256),
	}

	s.hub.Register <- client

	msgs, _ := s.repo.GetMessages(dialogID)
	for _, m := range msgs {
		client.Send <- m
	}

	go s.readPump(client)
	go s.writePump(client)
}

func (s *Server) readPump(c *chat.Client) {
	defer func() {
		s.hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg chat.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		msg.Username = c.ID
		msg.DialogID = c.DialogID

		if err := s.repo.SaveMessage(msg.DialogID, msg.Username, msg.Text); err != nil {
			continue
		}

		s.hub.Broadcast <- msg
	}
}

func (s *Server) writePump(c *chat.Client) {
	for msg := range c.Send {
		b, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		if err := c.Conn.WriteMessage(websocket.TextMessage, b); err != nil {
			return
		}
	}
}