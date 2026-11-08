
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"stopka/internal/decks"
)

type QuestionsHandler struct {
	repository *decks.Repository
}

func NewQuestionsHandler(repository *decks.Repository) *QuestionsHandler {
	return &QuestionsHandler{repository: repository}
}

func (h *QuestionsHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	deckID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	questions, err := h.repository.GetQuestions(r.Context(), deckID)
	if err != nil {
		log.Println("Ошибка получения вопросов:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"questions": questions,
	}); err != nil {
		log.Println("Ошибка отправки вопросов:", err)
	}
}
