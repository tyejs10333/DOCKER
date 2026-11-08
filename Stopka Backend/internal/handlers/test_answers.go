package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"stopka/internal/tests"
)

type TestAnswersHandler struct {
	repository *tests.Repository
}

func NewTestAnswersHandler(repository *tests.Repository) *TestAnswersHandler {
	return &TestAnswersHandler{repository: repository}
}

func (h *TestAnswersHandler) SaveAnswer(w http.ResponseWriter, r *http.Request) {
	attemptID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || attemptID <= 0 {
		http.Error(w, "Неверный ID попытки", http.StatusBadRequest)
		return
	}

	var request struct {
		QuestionID int64 `json:"question_id"`
		OptionID   int64 `json:"option_id"`
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

	if request.QuestionID <= 0 || request.OptionID <= 0 {
		http.Error(w, "Неверный ID вопроса или ответа", http.StatusBadRequest)
		return
	}

	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	answer, err := h.repository.SaveAnswer(
		r.Context(),
		attemptID,
		userID,
		request.QuestionID,
		request.OptionID,
	)

	if err != nil {
		switch {
		case errors.Is(err, tests.ErrAttemptNotFound):
			http.Error(w, "Попытка не найдена или уже завершена", http.StatusNotFound)
		case errors.Is(err, tests.ErrInvalidAnswer):
			http.Error(w, "Вопрос или ответ не подходит", http.StatusBadRequest)
		case errors.Is(err, tests.ErrAlreadyAnswered):
			http.Error(w, "На этот вопрос уже ответили", http.StatusConflict)
		default:
			log.Println("Ошибка сохранения ответа:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	correctOptionID, correctCount, wrongCount, err := h.repository.AnswerFeedback(r.Context(), attemptID, userID, request.QuestionID)
	if err != nil {
		log.Println("Ошибка получения результата ответа:", err)
		http.Error(w, "Ответ сохранён, но результат недоступен", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]any{
		"answer":            answer,
		"correct":           answer.IsCorrect,
		"correct_option_id": correctOptionID,
		"correct_count":     correctCount,
		"wrong_count":       wrongCount,
	}); err != nil {
		log.Println("Ошибка отправки ответа:", err)
	}
}
