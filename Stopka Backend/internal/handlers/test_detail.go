package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"stopka/internal/tests"
)

type TestDetailHandler struct {
	repository *tests.Repository
}

func NewTestDetailHandler(repository *tests.Repository) *TestDetailHandler {
	return &TestDetailHandler{repository: repository}
}

func (h *TestDetailHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	attemptID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || attemptID <= 0 {
		http.Error(w, "Неверный ID теста", http.StatusBadRequest)
		return
	}

	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	attempt, err := h.repository.GetFinishedAttempt(
		r.Context(),
		attemptID,
		userID,
	)
	if err != nil {
		if errors.Is(err, tests.ErrAttemptNotFound) {
			http.Error(w, "Тест не найден", http.StatusNotFound)
			return
		}

		log.Println("Ошибка получения результата теста:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	review, err := h.repository.GetAttemptReview(
		r.Context(),
		attemptID,
		userID,
	)
	if err != nil {
		log.Println("Ошибка получения разбора теста:", err)
		http.Error(w, "Не удалось получить разбор теста", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"attempt": attempt,
		"review":  review,
	}); err != nil {
		log.Println("Ошибка отправки результата теста:", err)
	}
}
