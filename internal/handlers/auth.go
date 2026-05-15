package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"

	"mini_chat/internal/middleware"
	"mini_chat/internal/models"
)

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := models.GetUserByUsername(r.Context(), username)
		if err != nil || user == nil || !models.CheckPasswordHash(password, user.Password) {
			tmpl := template.Must(template.ParseFiles("templates/html/login.html"))
			tmpl.Execute(w, map[string]string{"Error": "Неверное имя или пароль"})
			return
		}

		sessionID := generateSessionID()
		if err := models.CreateSession(r.Context(), sessionID, username); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
		})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/html/login.html"))
	tmpl.Execute(w, nil)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			tmpl := template.Must(template.ParseFiles("templates/html/register.html"))
			tmpl.Execute(w, map[string]string{"Error": "Заполните все поля"})
			return
		}

		existing, _ := models.GetUserByUsername(r.Context(), username)
		if existing != nil {
			tmpl := template.Must(template.ParseFiles("templates/html/register.html"))
			tmpl.Execute(w, map[string]string{"Error": "Пользователь уже существует"})
			return
		}

		hashed, err := models.HashPassword(password)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if err := models.CreateUser(r.Context(), username, hashed); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/html/register.html"))
	tmpl.Execute(w, nil)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserKey).(string)
	tmpl := template.Must(template.ParseFiles("templates/html/dashboard.html"))
	tmpl.Execute(w, username)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		models.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}