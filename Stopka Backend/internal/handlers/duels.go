package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DuelsHandler struct {
	db *pgxpool.Pool
}

func NewDuelsHandler(db *pgxpool.Pool) *DuelsHandler {
	return &DuelsHandler{db: db}
}

func getDuelUserID(r *http.Request) (int64, error) {
	id, ok := AuthenticatedUserID(r.Context())
	if !ok {
		return 0, errors.New("необходима авторизация")
	}
	return id, nil
}

func writeDuelJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Ошибка отправки JSON дуэли:", err)
	}
}

func newDuelCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	code := make([]byte, len(randomBytes))
	for i, b := range randomBytes {
		code[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(code), nil
}

func (h *DuelsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getDuelUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request struct {
		DeckID int64 `json:"deck_id"`
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

	if request.DeckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		log.Println("Ошибка начала транзакции дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var questionCount int

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM questions
		WHERE deck_id = $1
	`, request.DeckID).Scan(&questionCount)

	if err != nil {
		log.Println("Ошибка проверки вопросов дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if questionCount == 0 {
		http.Error(w, "В колоде нет вопросов", http.StatusBadRequest)
		return
	}

	var duelID int64
	var code string

	for attempt := 0; attempt < 5; attempt++ {
		code, err = newDuelCode()
		if err != nil {
			break
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO duels (
				code, deck_id, created_by, status
			)
			VALUES ($1, $2, $3, 'waiting')
			ON CONFLICT (code) DO NOTHING
			RETURNING id
		`, code, request.DeckID, userID).Scan(&duelID)

		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		break
	}

	if err != nil {
		log.Println("Ошибка создания дуэли:", err)
		http.Error(w, "Не удалось создать дуэль", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO duel_players (duel_id, user_id)
		VALUES ($1, $2)
	`, duelID, userID)

	if err != nil {
		log.Println("Ошибка добавления создателя дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO duel_questions (
			duel_id, question_id, position
		)
		SELECT $1, id, position
		FROM questions
		WHERE deck_id = $2
	`, duelID, request.DeckID)

	if err != nil {
		log.Println("Ошибка добавления вопросов дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("Ошибка сохранения дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeDuelJSON(w, http.StatusCreated, map[string]any{
		"duel": map[string]any{
			"id":         duelID,
			"code":       code,
			"deck_id":    request.DeckID,
			"created_by": userID,
			"status":     "waiting",
		},
	})
}

func (h *DuelsHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, err := getDuelUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request struct {
		Code string `json:"code"`
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

	if len(request.Code) != 8 {
		http.Error(w, "Неверный код дуэли", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var duelID, deckID, creatorID int64
	var status string

	err = tx.QueryRow(ctx, `
		SELECT id, deck_id, created_by, status
		FROM duels
		WHERE code = $1
		FOR UPDATE
	`, request.Code).Scan(
		&duelID, &deckID, &creatorID, &status,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Дуэль не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Ошибка поиска дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if creatorID == userID {
		http.Error(w, "Нельзя играть против себя", http.StatusConflict)
		return
	}

	if status != "waiting" {
		http.Error(w, "Дуэль уже началась или завершена", http.StatusConflict)
		return
	}

	var playerCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM duel_players
		WHERE duel_id = $1
	`, duelID).Scan(&playerCount)

	if err != nil {
		log.Println("Ошибка проверки участников:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if playerCount != 1 {
		http.Error(w, "Дуэль недоступна для присоединения", http.StatusConflict)
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO duel_players (duel_id, user_id)
		VALUES ($1, $2)
	`, duelID, userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "Вы уже участвуете в дуэли", http.StatusConflict)
			return
		}

		log.Println("Ошибка добавления игрока:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx, `
		UPDATE duels
		SET status = 'active', started_at = NOW()
		WHERE id = $1
	`, duelID)

	if err != nil {
		log.Println("Ошибка запуска дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("Ошибка сохранения участников:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeDuelJSON(w, http.StatusOK, map[string]any{
		"duel": map[string]any{
			"id":          duelID,
			"code":        request.Code,
			"deck_id":     deckID,
			"created_by":  creatorID,
			"opponent_id": userID,
			"status":      "active",
		},
	})
}
