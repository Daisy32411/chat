package main

import (
    "log"
    "net/http"

    "chat-app/internal/config"
    "chat-app/internal/db"
    "chat-app/internal/handler"
    "chat-app/internal/middleware"
    "chat-app/internal/repository"
    "chat-app/internal/service"
)

func main() {
    cfg := config.Load()

    database, err := db.Open(cfg.DBURL)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    userRepo := repository.NewUserRepo(database)
    sessionRepo := repository.NewSessionRepo(database)
    dialogRepo := repository.NewDialogRepo(database)
    messageRepo := repository.NewMessageRepo(database)

    authSvc := service.NewAuthService(userRepo, sessionRepo)
    chatSvc := service.NewChatService(dialogRepo, messageRepo, userRepo)

    authHandler := handler.NewAuthHandler(authSvc, cfg, database)
    chatHandler := handler.NewChatHandler(chatSvc)

    mux := http.NewServeMux()

    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/login", http.StatusFound)
    })
    mux.HandleFunc("/login", authHandler.LoginPage)
    mux.HandleFunc("/register", authHandler.RegisterPage)

    pageAuth := middleware.PageAuthMiddleware(database, cfg.CookieName, "/login")
    apiAuth := middleware.AuthMiddleware(database, cfg.CookieName)

    mux.Handle("/chat", pageAuth(http.HandlerFunc(authHandler.ChatPage)))
    mux.Handle("/api/me", apiAuth(http.HandlerFunc(authHandler.Me)))

    mux.HandleFunc("/api/auth/login", authHandler.Login)
    mux.HandleFunc("/api/auth/register", authHandler.Register)
    mux.Handle("/api/auth/logout", apiAuth(http.HandlerFunc(authHandler.Logout)))

    mux.Handle("/api/dialogs", apiAuth(http.HandlerFunc(chatHandler.ListDialogs)))
    mux.Handle("/api/users/search", apiAuth(http.HandlerFunc(chatHandler.SearchUsers)))
    mux.Handle("/api/dialogs/open", apiAuth(http.HandlerFunc(chatHandler.OpenDialog)))
    mux.Handle("/api/messages", apiAuth(http.HandlerFunc(chatHandler.ListMessages)))
    mux.Handle("/api/messages/send", apiAuth(http.HandlerFunc(chatHandler.SendMessage)))

    log.Printf("server started on :%s", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
        log.Fatal(err)
    }
}
