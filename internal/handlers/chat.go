package handlers

import (
	"context"
	"encoding/json"
	"log"
	"mini_chat/internal/middleware"
	"mini_chat/internal/models"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn     *websocket.Conn
	userID   int
	username string
	send     chan []byte
	ctx      context.Context
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var hub = &Hub{
	clients:    make(map[*Client]bool),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) sendToUser(username string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.username == username {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
			break
		}
	}
}

func RunHub() {
	go hub.run()
}

func ChatWebSocket(w http.ResponseWriter, r *http.Request) {
	usernameVal := r.Context().Value(middleware.UserKey)
	if usernameVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	username := usernameVal.(string)

	userID, err := models.GetUserIDByUsername(r.Context(), username)
	if err != nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade failed:", err)
		return
	}

	client := &Client{
		conn:     conn,
		userID:   userID,
		username: username,
		send:     make(chan []byte, 256),
		ctx:      r.Context(),
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
		hub.unregister <- c
	}()
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		hub.unregister <- c
	}()
	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg struct {
			To   string `json:"to"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}
		if msg.To == "" || msg.Text == "" {
			continue
		}
		// Создаём новый контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = models.SaveMessage(ctx, c.userID, msg.To, msg.Text)
		cancel()
		if err != nil {
			log.Printf("Save message error: %v", err)
			continue
		}
		// Отправляем обратно отправителю
		outMsg := map[string]interface{}{
			"from": c.username,
			"to":   msg.To,
			"text": msg.Text,
		}
		outBytes, _ := json.Marshal(outMsg)
		c.send <- outBytes
		// Отправляем получателю, если онлайн
		hub.sendToUser(msg.To, outBytes)
	}
}

// REST API handlers

func GetDialogsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserKey).(string)
	userID, err := models.GetUserIDByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	dialogs, err := models.GetDialogs(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dialogs)
}

func SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}
	username := r.Context().Value(middleware.UserKey).(string)
	userID, _ := models.GetUserIDByUsername(r.Context(), username)
	results, err := models.SearchUsers(r.Context(), userID, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	otherUser := r.URL.Query().Get("with")
	if otherUser == "" {
		http.Error(w, "missing 'with' param", http.StatusBadRequest)
		return
	}
	username := r.Context().Value(middleware.UserKey).(string)
	userID, _ := models.GetUserIDByUsername(r.Context(), username)
	messages, err := models.GetMessagesBetween(r.Context(), userID, otherUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}