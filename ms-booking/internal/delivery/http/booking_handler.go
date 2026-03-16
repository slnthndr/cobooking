package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/slnt/cobooking/ms-booking/internal/domain"
)

type BookingHandler struct {
	service domain.BookingService
}

func NewBookingHandler(r chi.Router, service domain.BookingService) {
	handler := &BookingHandler{service: service}
	r.Post("/api/v1/bookings/{placeId}", handler.Create)
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим PlaceID из URL
	placeIDStr := chi.URLParam(r, "placeId")
	placeID, err := strconv.Atoi(placeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid place ID")
		return
	}

	// 2. Парсим UserID из заголовка X-User-Id (Его должен прислать API Gateway после проверки токена)
	userIDStr := r.Header.Get("X-User-Id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid User ID")
		return
	}

	// 3. Читаем тело запроса (входные даты в ISO 8601 / RFC3339)
	var req struct {
		StartTime string `json:"bookingStartTime"`
		EndTime   string `json:"bookingEndTime"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON format")
		return
	}

	// Конвертируем строки во время (обязательно с 'Z' на конце для парсинга)
	start, errStart := time.Parse(time.RFC3339, req.StartTime)
	end, errEnd := time.Parse(time.RFC3339, req.EndTime)
	if errStart != nil || errEnd != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Wrong booking time format")
		return
	}

	// Проверяем логику (End > Start)
	if !end.After(start) {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "bookingEndTime must be after bookingStartTime")
		return
	}

	// 4. Передаем в сервис
	bookingReq := domain.CreateBookingRequest{
		PlaceID:   placeID,
		UserID:    userID,
		StartTime: start,
		EndTime:   end,
	}

	booking, err := h.service.CreateBooking(r.Context(), bookingReq)
	if err != nil {
		if err.Error() == "SLOT_ALREADY_RESERVED" {
			writeError(w, http.StatusConflict, "SLOT_ALREADY_RESERVED", "This slot is already reserved")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create booking")
		return
	}

	// Успешный ответ (ТЗ стр. 54: HTTP 201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

func writeError(w http.ResponseWriter, status int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": errCode, "message": msg})
}
