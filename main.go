package main

import (
	"log"
	"net/http"

	"chat/internal/auth"
	"chat/internal/chat"
	"chat/internal/db"
	"chat/internal/dialogs"
	"chat/internal/middleware"
	"chat/internal/users"
	httpTransport "chat/internal/transport/http"
	ws "chat/internal/transport/websocket"
)

func main() {
	database, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo)
	authHandler := httpTransport.NewHandler(authService)

	dialogRepo := dialogs.NewRepository(database)
	dialogService := dialogs.NewService(dialogRepo)
	dialogHandler := httpTransport.NewDialogHandler(dialogService)

	chatRepo := chat.NewRepository(database)
	messagesHandler := httpTransport.NewMessagesHandler(chatRepo, dialogRepo)

	usersRepo := users.NewRepository(database)
	usersHandler := httpTransport.NewUsersHandler(usersRepo)

	hub := chat.NewHub()
	go hub.Run()

	wsServer := ws.NewServer(hub, chatRepo, dialogRepo)

	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/me", middleware.AuthMiddleware(authHandler.Me))

	http.HandleFunc("/dialogs", middleware.AuthMiddleware(dialogHandler.GetDialogs))
	http.HandleFunc("/dialogs/create", middleware.AuthMiddleware(dialogHandler.CreateDialog))

	http.HandleFunc("/messages", middleware.AuthMiddleware(messagesHandler.GetMessages))
	http.HandleFunc("/users/search", middleware.AuthMiddleware(usersHandler.SearchUsers))

	http.HandleFunc("/ws", wsServer.ServeWs)

	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("server started :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}