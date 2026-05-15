package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "chat-app/internal/middleware"
    "chat-app/internal/service"
)

type ChatHandler struct {
    svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
    return &ChatHandler{svc: svc}
}

func (h *ChatHandler) ListDialogs(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.CurrentUser(r)
    if !ok {
        errorJSON(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    items, err := h.svc.ListDialogs(r.Context(), user.ID)
    if err != nil {
        errorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"dialogs": items})
}

func (h *ChatHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
    if _, ok := middleware.CurrentUser(r); !ok {
        errorJSON(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    query := r.URL.Query().Get("q")
    users, err := h.svc.UsersSearch(r.Context(), query)
    if err != nil {
        errorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

type openDialogRequest struct {
    UserID int64 `json:"user_id"`
}

func (h *ChatHandler) OpenDialog(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.CurrentUser(r)
    if !ok {
        errorJSON(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    var req openDialogRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errorJSON(w, http.StatusBadRequest, "bad json")
        return
    }
    dialogID, err := h.svc.OpenDialog(r.Context(), user.ID, req.UserID)
    if err != nil {
        errorJSON(w, http.StatusBadRequest, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"dialog_id": dialogID})
}

func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
    if _, ok := middleware.CurrentUser(r); !ok {
        errorJSON(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    dialogID, err := strconv.ParseInt(r.URL.Query().Get("dialog_id"), 10, 64)
    if err != nil || dialogID <= 0 {
        errorJSON(w, http.StatusBadRequest, "invalid dialog_id")
        return
    }
    messages, err := h.svc.ListMessages(r.Context(), dialogID)
    if err != nil {
        errorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

type sendMessageRequest struct {
    DialogID int64  `json:"dialog_id"`
    Text     string `json:"text"`
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.CurrentUser(r)
    if !ok {
        errorJSON(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    var req sendMessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errorJSON(w, http.StatusBadRequest, "bad json")
        return
    }
    msg, err := h.svc.SendMessage(r.Context(), req.DialogID, user.ID, req.Text)
    if err != nil {
        errorJSON(w, http.StatusBadRequest, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"message": msg})
}
