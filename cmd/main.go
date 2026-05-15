package main

import (
	"log"
	"mini_chat/config"
	"mini_chat/internal/db"
	"mini_chat/internal/handlers"
	"mini_chat/internal/middleware"
	"net/http"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

	// Запускаем WebSocket hub
	go handlers.RunHub()

	// Статические файлы (CSS, JS) из папки templates
	// Теперь URL вида /static/css/login.css будут искать файл в ./templates/css/login.css
	fs := http.FileServer(http.Dir("./templates"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/dashboard", middleware.AuthMiddleware(handlers.DashboardHandler))
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Маршруты чата
	http.HandleFunc("/ws", middleware.AuthMiddleware(handlers.ChatWebSocket))
	http.HandleFunc("/api/dialogs", middleware.AuthMiddleware(handlers.GetDialogsHandler))
	http.HandleFunc("/api/search", middleware.AuthMiddleware(handlers.SearchUsersHandler))
	http.HandleFunc("/api/messages", middleware.AuthMiddleware(handlers.GetMessagesHandler))

	log.Println("Server started on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}