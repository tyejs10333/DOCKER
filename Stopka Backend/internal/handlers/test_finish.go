package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"stopka/internal/decks"
	"stopka/internal/tests"
)

type TestFinishHandler struct {
	repository     *tests.Repository
	deckRepository *decks.Repository
}

func NewTestFinishHandler(
	repository *tests.Repository,
	deckRepository *decks.Repository,
) *TestFinishHandler {
	return &TestFinishHandler{
		repository:     repository,
		deckRepository: deckRepository,
	}
}

func (h *TestFinishHandler) FinishAttempt(w http.ResponseWriter, r *http.Request) {
	attemptID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || attemptID <= 0 {
		http.Error(w, "Неверный ID попытки", http.StatusBadRequest)
		return
	}

	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	attempt, err := h.repository.FinishAttempt(
		r.Context(),
		attemptID,
		userID,
	)

	if err != nil {
		switch {
		case errors.Is(err, tests.ErrAttemptNotFound):
			http.Error(w, "Попытка не найдена", http.StatusNotFound)

		case errors.Is(err, tests.ErrAttemptIncomplete):
			http.Error(w, "Сначала ответьте на все вопросы", http.StatusConflict)

		case errors.Is(err, tests.ErrAttemptFinished):
			http.Error(w, "Тест уже завершён", http.StatusConflict)

		default:
			log.Println("Ошибка завершения теста:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	review, err := h.repository.GetAttemptReview(
		r.Context(),
		attemptID,
		userID,
	)
	if err != nil {
		log.Println("Ошибка получения разбора теста:", err)
		http.Error(w, "Тест завершён, но не удалось получить разбор", http.StatusInternalServerError)
		return
	}

	deckTitle := ""

	userDecks, err := h.deckRepository.GetAllForUser(
		r.Context(),
		userID,
	)
	if err != nil {
		log.Println("Ошибка получения названия колоды:", err)
	} else {
		for _, deck := range userDecks {
			if deck.ID == attempt.DeckID {
				deckTitle = deck.Title
				break
			}
		}
	}

	response := map[string]any{
		"attempt": map[string]any{
			"id":          attempt.ID,
			"deck_id":     attempt.DeckID,
			"user_id":     attempt.UserID,
			"status":      attempt.Status,
			"started_at":  attempt.StartedAt,
			"finished_at": attempt.FinishedAt,
			"correct":     attempt.Correct,
			"total":       attempt.Total,
			"percent":     attempt.Percent,
			"deck": map[string]any{
				"id":    attempt.DeckID,
				"title": deckTitle,
			},
		},
		"review": review,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Ошибка отправки результата теста:", err)
	}
}
