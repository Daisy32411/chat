package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"mini_chat/config"
	"mini_chat/internal/db"
	"mini_chat/internal/handlers"
	authMiddleware "mini_chat/internal/middleware"
)

func setupRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fs := http.FileServer(http.Dir("./templates"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", handlers.IndexHandler)
	r.Get("/login", handlers.LoginHandler)
	r.Post("/login", handlers.LoginHandler)
	r.Get("/register", handlers.RegisterHandler)
	r.Post("/register", handlers.RegisterHandler)
	r.Get("/logout", handlers.LogoutHandler)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)
		r.Get("/dashboard", handlers.DashboardHandler)
		r.Get("/ws", handlers.ChatWebSocket)
		r.Get("/api/dialogs", handlers.GetDialogsHandler)
		r.Get("/api/search", handlers.SearchUsersHandler)
		r.Get("/api/messages", handlers.GetMessagesHandler)
	})

	return r
}

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

	go handlers.RunHub()

	r := setupRouter()

	log.Printf("Server started on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}