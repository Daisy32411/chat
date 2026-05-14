package chat

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client

	repo	   *Repository
}

func NewHub(repo *Repository) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		repo: 		repo,
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
			_ = h.repo.SaveMessage(msg)

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