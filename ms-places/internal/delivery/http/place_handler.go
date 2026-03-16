package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"fmt"
	"os"
	"io"
	"path/filepath"


	"github.com/slnt/cobooking/ms-places/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/slnt/cobooking/ms-places/internal/domain"
)

type PlaceHandler struct {
	service domain.PlaceService
	s3storage *storage.S3Storage
}

func NewPlaceHandler(r chi.Router, service domain.PlaceService, s3 *storage.S3Storage) {
    handler := &PlaceHandler{
        service: service,
        s3storage: s3,
    }

	// Роуты из ТЗ
	r.Get("/api/v1/places/{placeId}", handler.GetPlace)
	r.Get("/api/v1/search", handler.Search)
	r.Post("/api/v1/places/{placeId}/image", handler.UploadImage)
}

func (h *PlaceHandler) GetPlace(w http.ResponseWriter, r *http.Request) {
	placeIDStr := chi.URLParam(r, "placeId")
	placeID, err := strconv.Atoi(placeIDStr)
	if err != nil {
		fmt.Println("Error in GetPlace:", err)
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid place ID")
		return
	}

	place, err := h.service.GetPlace(r.Context(), placeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Place not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 минут кэширования на стороне клиента
	
	json.NewEncoder(w).Encode(place)
}

func (h *PlaceHandler) Search(w http.ResponseWriter, r *http.Request) {
	// Читаем query параметры
	country := r.URL.Query().Get("country")
	city := r.URL.Query().Get("city")
	
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	offset := (page - 1) * limit

	places, err := h.service.SearchPlaces(r.Context(), country, city, limit, offset)
	if err != nil {
		fmt.Println("Error in Search:", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search places")
		return
	}

	// Отвечаем по формату ТЗ (со списком и пагинацией)
	response := map[string]interface{}{
		"places": places,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, status int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errCode,
		"message": msg,
	})
}

// UploadImage принимает файл через multipart/form-data и кидает в S3
func (h *PlaceHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	placeIDStr := chi.URLParam(r, "placeId")

	// Ограничиваем размер файла до 5 МБ
	r.ParseMultipartForm(5 << 20)
	
	file, handler, err := r.FormFile("image") // "image" - имя поля в Postman
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Failed to get image from request")
		return
	}
	defer file.Close()

	// Сохраняем файл во временную директорию ОС
	tempFile, err := os.CreateTemp("", "upload-*.jpg")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create temp file")
		return
	}
	defer os.Remove(tempFile.Name()) // Удалим временный файл после загрузки в S3

	io.Copy(tempFile, file)

	// Генерируем уникальное имя для S3: place_{id}_{filename}
	objectName := fmt.Sprintf("place_%s_%s", placeIDStr, filepath.Base(handler.Filename))

	// Загружаем в S3
	imageURL, err := h.s3storage.UploadImage(r.Context(), objectName, tempFile.Name(), handler.Header.Get("Content-Type"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload image to S3")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Image uploaded successfully",
		"url":     imageURL,
	})
}
