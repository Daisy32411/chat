package chat

import "github.com/gorilla/websocket"

type Client struct {
	ID   	  string
	DialogID  int
	Conn 	  *websocket.Conn
	Send 	  chan Message
}