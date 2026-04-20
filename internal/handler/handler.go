package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RomanGolovinn/urlshortener/internal/repository"
	"github.com/RomanGolovinn/urlshortener/internal/service"
)

type Handler struct {
	service *service.Service
	host    string
}

func NewHandler(service *service.Service, host string) *Handler {
	return &Handler{
		service: service,
		host:    host,
	}
}

type requestDTO struct {
	URL string `json:"url"`
}

type responseDTO struct {
	ShortURL string `json:"short_url"`
}

type errorResponseDTO struct {
	Error string `json:"error"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req requestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.URL == "" {
		h.writeError(w, "url не может быть пустым", http.StatusBadRequest)
		return
	}

	shortCode, err := h.service.ShortenURL(r.Context(), req.URL)
	if err != nil {
		h.writeError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	resp := responseDTO{
		ShortURL: h.host + shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("code")
	if shortCode == "" {
		h.writeError(w, "код не указан", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.ProcessRedirect(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.writeError(w, "ссылка не найдена", http.StatusNotFound)
			return
		}
		h.writeError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponseDTO{Error: message})
}
