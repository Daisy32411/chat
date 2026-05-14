package ws

import (
	"encoding/json"
	"net/http"

	"chat/internal/chat"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &chat.Client{
		Conn: conn,
		Send: make(chan chat.Message, 256),
	}

	hub.Register <- client

	go readPump(hub, client)
	go writePump(client)
}

func readPump(hub *chat.Hub, c *chat.Client) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg chat.Message
		json.Unmarshal(data, &msg)

		hub.Broadcast <- msg
	}
}

func writePump(c *chat.Client) {
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