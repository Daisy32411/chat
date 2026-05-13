package ws

import (
	"net/http"

	"chat/internal/model"
	"chat/internal/server"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(hub *server.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &model.Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client

	go readPump(hub, client)
	go writePump(client)
}

func readPump(hub *server.Hub, c *model.Client) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		hub.Broadcast <- msg
	}
}

func writePump(c *model.Client) {
	for msg := range c.Send {
		c.Conn.WriteMessage(1, msg)
	}
}