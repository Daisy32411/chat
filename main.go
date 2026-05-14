package main

import (
	"log"
	"net/http"

	"chat/internal/db"
	"chat/internal/auth"
	httpTransport "chat/internal/transport/http"
	ws "chat/internal/transport/websocket"
	"chat/internal/chat"
)

func main() {
	database, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	repo := auth.NewRepository(database)
	service := auth.NewService(repo)
	handler := httpTransport.NewHandler(service)

	hub := chat.NewHub()
	go hub.Run()

	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/login", handler.Login)

	http.HandleFunc("/me", handler.Me)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("server started :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}