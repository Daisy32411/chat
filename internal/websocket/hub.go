package websocket

// This package is kept only as a placeholder for a future WebSocket
// implementation. The current app uses polling for message updates.

type Event struct {
    Type     string
    DialogID int64
    Payload  any
}

type Hub struct{}

func NewHub() *Hub { return &Hub{} }

func (h *Hub) Broadcast(dialogID int64, event Event) {}
