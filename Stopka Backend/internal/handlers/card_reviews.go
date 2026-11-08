package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CardReviewsHandler struct {
	db *pgxpool.Pool
}

func NewCardReviewsHandler(db *pgxpool.Pool) *CardReviewsHandler {
	return &CardReviewsHandler{db: db}
}

func (h *CardReviewsHandler) SaveReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var request struct {
		CardID int64  `json:"card_id"`
		Grade  string `json:"grade"`
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 4096))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "Ожидается один JSON-объект", http.StatusBadRequest)
		return
	}

	request.Grade = strings.TrimSpace(request.Grade)

	if request.CardID <= 0 ||
		(request.Grade != "again" &&
			request.Grade != "hard" &&
			request.Grade != "good") {
		http.Error(w, "Неверная карточка или оценка", http.StatusBadRequest)
		return
	}

	var reviewedAt string

	err := h.db.QueryRow(r.Context(), `
		INSERT INTO card_reviews (user_id, card_id, grade)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, card_id)
		DO UPDATE SET
			grade = EXCLUDED.grade,
			reviewed_at = NOW()
		RETURNING reviewed_at::text
	`, userID, request.CardID, request.Grade).Scan(&reviewedAt)

	if err != nil {
		log.Printf(
			"Ошибка сохранения карточки: user_id=%d card_id=%d: %v",
			userID, request.CardID, err,
		)
		http.Error(w, "Не удалось сохранить оценку карточки", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"review": map[string]any{
			"card_id":     request.CardID,
			"grade":       request.Grade,
			"reviewed_at": reviewedAt,
		},
	}); err != nil {
		log.Println("Ошибка отправки результата карточки:", err)
	}
}
