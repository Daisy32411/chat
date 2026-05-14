package chat

type Message struct {
	Type string `json:"type"`
	Username string `json:"username"`
	Text string `json:"text"`
}