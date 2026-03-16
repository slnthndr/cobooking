package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/slnt/cobooking/ms-auth/internal/domain"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(r chi.Router, service domain.AuthService) {
	handler := &AuthHandler{service: service}

	// Регистрируем роуты в соответствии с ТЗ
	r.Post("/api/v1/users", handler.Register)
	r.Post("/api/v1/auth/login", handler.Login)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON body")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusConflict, "CONFLICT", "Email уже зарегистрирован")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON body")
		return
	}

	tokens, err := h.service.Login(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный email или пароль")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}

// Хелпер для отправки ошибок в формате ТЗ
func writeError(w http.ResponseWriter, status int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errCode,
		"message": msg,
	})
}
