package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/slnt/cobooking/ms-payments/internal/domain"
)

type PaymentHandler struct {
	service domain.PaymentService
}

func NewPaymentHandler(r chi.Router, service domain.PaymentService) {
	h := &PaymentHandler{service: service}
	r.Post("/api/v1/payments/pay", h.Pay)
	r.Post("/api/v1/payments/webhook", h.Webhook)
}

func (h *PaymentHandler) Pay(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-Id")
	userID, _ := strconv.Atoi(userIDStr)

	var req domain.PayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON")
		return
	}
	req.UserID = userID

	payment, err := h.service.Pay(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create payment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	var payload domain.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON")
		return
	}

	if err := h.service.ProcessWebhook(r.Context(), payload); err != nil {
		fmt.Println("Error in ProcessWebhook:", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Webhook processed"}`))
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code, "message": msg})
}
