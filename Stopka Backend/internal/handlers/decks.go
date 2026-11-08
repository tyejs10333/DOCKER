package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"stopka/internal/decks"
)

type DecksHandler struct {
	repository *decks.Repository
}

func NewDecksHandler(repository *decks.Repository) *DecksHandler {
	return &DecksHandler{repository: repository}
}

func (h *DecksHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	result, err := h.repository.GetAllForUser(r.Context(), userID)
	if err != nil {
		log.Printf("Ошибка получения колод: user_id=%d: %v", userID, err)
		http.Error(w, "Не удалось получить колоды", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"decks": result,
	}); err != nil {
		log.Printf("Ошибка отправки колод: %v", err)
	}
}
