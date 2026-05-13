package main

import (
	"log"
	"net/http"

	"chat/internal/server"
	"chat/internal/ws"
)

func main() {
	// создаём hub (центр системы)
	hub := server.NewHub()

	// запускаем его в отдельной горутине
	go hub.Run()

	// HTTP endpoint для websocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})
	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("server started on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}