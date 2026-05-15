package chat

type Message struct {
	ID       int    `json:"id"`
	DialogID int    `json:"dialog_id"`
	Username string `json:"username"`
	Text     string `json:"text"`
}