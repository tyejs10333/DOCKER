
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"stopka/internal/decks"
)

type AnswersHandler struct {
	repository *decks.Repository
}

func NewAnswersHandler(repository *decks.Repository) *AnswersHandler {
	return &AnswersHandler{repository: repository}
}

func (h *AnswersHandler) CheckAnswer(w http.ResponseWriter, r *http.Request) {
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

	isCorrect, err := h.repository.CheckAnswer(
		r.Context(),
		request.QuestionID,
		request.OptionID,
	)
	if err != nil {
		log.Println("Ошибка проверки ответа:", err)
		http.Error(w, "Не удалось проверить ответ", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"is_correct": isCorrect,
	}); err != nil {
		log.Println("Ошибка отправки результата:", err)
	}
}
