package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"stopka/internal/tests"
)

type TestHistoryHandler struct {
	repository *tests.Repository
}

func NewTestHistoryHandler(repository *tests.Repository) *TestHistoryHandler {
	return &TestHistoryHandler{repository: repository}
}

func (h *TestHistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	history, err := h.repository.GetUserTestHistory(r.Context(), userID)
	if err != nil {
		log.Println("Ошибка получения истории тестов:", err)
		http.Error(w, "Не удалось получить историю тестов", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"attempts": history,
	}); err != nil {
		log.Println("Ошибка отправки истории тестов:", err)
	}
}
