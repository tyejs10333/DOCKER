
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"stopka/internal/decks"
)

type CardsHandler struct {
	repository *decks.Repository
}

func NewCardsHandler(repository *decks.Repository) *CardsHandler {
	return &CardsHandler{repository: repository}
}

func (h *CardsHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	deckID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "all"
	}
	if mode != "all" && mode != "new" && mode != "weak" {
		http.Error(w, "Неверный режим карточек", http.StatusBadRequest)
		return
	}

	deckList, err := h.repository.GetAll(r.Context())
	if err != nil {
		log.Printf("Ошибка поиска колоды: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	var selected *decks.Deck
	for i := range deckList {
		if deckList[i].ID == deckID {
			selected = &deckList[i]
			break
		}
	}

	if selected == nil {
		http.Error(w, "Колода не найдена", http.StatusNotFound)
		return
	}

	cards, err := h.repository.GetStudyCards(r.Context(), deckID, userID, mode)
	if err != nil {
		log.Printf("Ошибка карточек: deck_id=%d user_id=%d mode=%s: %v",
			deckID, userID, mode, err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"deck": map[string]any{
			"id":    selected.ID,
			"title": selected.Title,
		},
		"cards": cards,
	}); err != nil {
		log.Printf("Ошибка отправки карточек: %v", err)
	}
}
