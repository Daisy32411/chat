package server

import "chat/internal/model"

type Hub struct {
	Clients    map[*model.Client]bool
	Broadcast  chan []byte
	Register   chan *model.Client
	Unregister chan *model.Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*model.Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *model.Client),
		Unregister: make(chan *model.Client),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.Register:
			h.Clients[client] = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}

		case msg := <-h.Broadcast:
			for c := range h.Clients {
				select {
				case c.Send <- msg:
				default:
					close(c.Send)
					delete(h.Clients, c)
				}
			}
		}
	}
}