package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"stopka/internal/decks"
	"stopka/internal/tests"
)

type TestAttemptsHandler struct {
	repository     *tests.Repository
	deckRepository *decks.Repository
}

func NewTestAttemptsHandler(
	repository *tests.Repository,
	deckRepository *decks.Repository,
) *TestAttemptsHandler {
	return &TestAttemptsHandler{
		repository:     repository,
		deckRepository: deckRepository,
	}
}

func (h *TestAttemptsHandler) CreateAttempt(w http.ResponseWriter, r *http.Request) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var request struct {
		DeckID int64 `json:"deck_id"`
		UserID int64 `json:"user_id"`
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 4096))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "Ожидается один JSON-объект", http.StatusBadRequest)
		return
	}

	if request.DeckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	questions, err := h.deckRepository.GetQuestions(
		r.Context(),
		request.DeckID,
	)
	if err != nil {
		log.Println("Ошибка получения вопросов:", err)
		http.Error(w, "Не удалось получить вопросы", http.StatusInternalServerError)
		return
	}

	if len(questions) == 0 {
		http.Error(w, "В выбранной колоде нет вопросов", http.StatusBadRequest)
		return
	}

	attempt, err := h.repository.CreateAttempt(
		r.Context(),
		userID,
		request.DeckID,
	)
	if err != nil {
		log.Println("Ошибка создания попытки:", err)
		http.Error(w, "Не удалось создать попытку", http.StatusBadRequest)
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
			if deck.ID == request.DeckID {
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
		"questions": questions,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Ошибка отправки попытки:", err)
	}
}
