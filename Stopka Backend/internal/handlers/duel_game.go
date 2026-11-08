package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func (h *DuelsHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	userID, err := getDuelUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	duelID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || duelID <= 0 {
		http.Error(w, "Неверный ID дуэли", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var code, status, deckTitle string
	var deckID int64
	var winnerID *int64

	err = h.db.QueryRow(ctx, `
		SELECT d.code, d.status, d.deck_id,
		       decks.title, d.winner_user_id
		FROM duels d
		JOIN decks ON decks.id = d.deck_id
		WHERE d.id = $1
		  AND EXISTS (
		      SELECT 1 FROM duel_players
		      WHERE duel_id = d.id AND user_id = $2
		  )
	`, duelID, userID).Scan(
		&code, &status, &deckID, &deckTitle, &winnerID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Дуэль не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Ошибка получения дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	playerRows, err := h.db.Query(ctx, `
		SELECT user_id, score, answered
		FROM duel_players
		WHERE duel_id = $1
		ORDER BY joined_at, user_id
	`, duelID)
	if err != nil {
		log.Println("Ошибка получения игроков:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	players := make([]map[string]any, 0, 2)
	for playerRows.Next() {
		var id int64
		var score, answered int

		if err := playerRows.Scan(&id, &score, &answered); err != nil {
			playerRows.Close()
			http.Error(w, "Ошибка чтения игроков", http.StatusInternalServerError)
			return
		}

		players = append(players, map[string]any{
			"user_id":  id,
			"score":    score,
			"answered": answered,
		})
	}
	err = playerRows.Err()
	playerRows.Close()
	if err != nil {
		log.Println("Ошибка чтения игроков:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	questionRows, err := h.db.Query(ctx, `
		SELECT q.id, q.text, qo.id, qo.text
		FROM duel_questions dq
		JOIN questions q ON q.id = dq.question_id
		LEFT JOIN question_options qo ON qo.question_id = q.id
		WHERE dq.duel_id = $1
		ORDER BY dq.position, q.id, qo.position, qo.id
	`, duelID)
	if err != nil {
		log.Println("Ошибка получения вопросов дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	type option struct {
		ID   int64  `json:"id"`
		Text string `json:"text"`
	}
	type question struct {
		ID      int64    `json:"id"`
		Text    string   `json:"text"`
		Options []option `json:"options"`
	}

	questions := make([]question, 0)
	indexes := make(map[int64]int)

	for questionRows.Next() {
		var questionID int64
		var questionText string
		var optionID *int64
		var optionText *string

		if err := questionRows.Scan(
			&questionID, &questionText, &optionID, &optionText,
		); err != nil {
			questionRows.Close()
			http.Error(w, "Ошибка чтения вопросов", http.StatusInternalServerError)
			return
		}

		index, exists := indexes[questionID]
		if !exists {
			index = len(questions)
			indexes[questionID] = index
			questions = append(questions, question{
				ID:      questionID,
				Text:    questionText,
				Options: make([]option, 0),
			})
		}

		if optionID != nil && optionText != nil {
			questions[index].Options = append(
				questions[index].Options,
				option{ID: *optionID, Text: *optionText},
			)
		}
	}
	err = questionRows.Err()
	questionRows.Close()
	if err != nil {
		log.Println("Ошибка чтения вопросов:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	answerRows, err := h.db.Query(ctx, `
		SELECT question_id, option_id, is_correct
		FROM duel_answers
		WHERE duel_id = $1 AND user_id = $2
	`, duelID, userID)
	if err != nil {
		log.Println("Ошибка получения ответов:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	answers := make([]map[string]any, 0)
	for answerRows.Next() {
		var questionID, optionID int64
		var correct bool

		if err := answerRows.Scan(&questionID, &optionID, &correct); err != nil {
			answerRows.Close()
			http.Error(w, "Ошибка чтения ответов", http.StatusInternalServerError)
			return
		}

		answers = append(answers, map[string]any{
			"question_id": questionID,
			"option_id":   optionID,
			"correct":     correct,
		})
	}
	err = answerRows.Err()
	answerRows.Close()
	if err != nil {
		log.Println("Ошибка чтения ответов:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeDuelJSON(w, http.StatusOK, map[string]any{
		"duel": map[string]any{
			"id":     duelID,
			"code":   code,
			"status": status,
			"deck": map[string]any{
				"id":    deckID,
				"title": deckTitle,
			},
			"winner_user_id": winnerID,
		},
		"players":   players,
		"questions": questions,
		"answers":   answers,
	})
}

func (h *DuelsHandler) Answer(w http.ResponseWriter, r *http.Request) {
	userID, err := getDuelUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	duelID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || duelID <= 0 {
		http.Error(w, "Неверный ID дуэли", http.StatusBadRequest)
		return
	}

	var request struct {
		QuestionID int64 `json:"question_id"`
		OptionID   int64 `json:"option_id"`
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

	if request.QuestionID <= 0 || request.OptionID <= 0 {
		http.Error(w, "Неверный вопрос или ответ", http.StatusBadRequest)
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

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM duels WHERE id = $1 FOR UPDATE
	`, duelID).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Дуэль не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Ошибка блокировки дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if status != "active" {
		http.Error(w, "Дуэль сейчас не активна", http.StatusConflict)
		return
	}

	var playerExists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM duel_players
			WHERE duel_id = $1 AND user_id = $2
		)
	`, duelID, userID).Scan(&playerExists)

	if err != nil || !playerExists {
		http.Error(w, "Вы не участвуете в дуэли", http.StatusForbidden)
		return
	}

	var correct bool
	err = tx.QueryRow(ctx, `
		SELECT qo.is_correct
		FROM duel_questions dq
		JOIN question_options qo
		  ON qo.question_id = dq.question_id
		WHERE dq.duel_id = $1
		  AND dq.question_id = $2
		  AND qo.id = $3
	`, duelID, request.QuestionID, request.OptionID).Scan(&correct)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Вопрос или вариант ответа не подходит", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Println("Ошибка проверки ответа:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	var alreadyAnswered bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM duel_answers
			WHERE duel_id = $1 AND user_id = $2 AND question_id = $3
		)
	`, duelID, userID, request.QuestionID).Scan(&alreadyAnswered)

	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if alreadyAnswered {
		http.Error(w, "На этот вопрос уже ответили", http.StatusConflict)
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO duel_answers (
			duel_id, user_id, question_id, option_id, is_correct
		)
		VALUES ($1, $2, $3, $4, $5)
	`, duelID, userID, request.QuestionID, request.OptionID, correct)

	if err != nil {
		log.Println("Ошибка сохранения ответа дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	scoreIncrement := 0
	if correct {
		scoreIncrement = 1
	}

	_, err = tx.Exec(ctx, `
		UPDATE duel_players
		SET answered = answered + 1,
		    score = score + $3
		WHERE duel_id = $1 AND user_id = $2
	`, duelID, userID, scoreIncrement)

	if err != nil {
		log.Println("Ошибка обновления очков:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	var totalQuestions, completedPlayers int

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM duel_questions WHERE duel_id = $1
	`, duelID).Scan(&totalQuestions)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM duel_players
		WHERE duel_id = $1 AND answered = $2
	`, duelID, totalQuestions).Scan(&completedPlayers)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	finished := completedPlayers == 2 && totalQuestions > 0

	if finished {
		var winnerID *int64

		type result struct {
			ID    int64
			Score int
		}
		var results []result

		rows, err := tx.Query(ctx, `
			SELECT user_id, score
			FROM duel_players
			WHERE duel_id = $1
			ORDER BY score DESC, user_id
		`, duelID)
		if err != nil {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			var item result
			if err := rows.Scan(&item.ID, &item.Score); err != nil {
				rows.Close()
				http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
				return
			}
			results = append(results, item)
		}
		readErr := rows.Err()
		rows.Close()
		if readErr != nil || len(results) != 2 {
			http.Error(w, "Ошибка результатов дуэли", http.StatusInternalServerError)
			return
		}

		if results[0].Score > results[1].Score {
			winnerID = &results[0].ID
		}

		_, err = tx.Exec(ctx, `
			UPDATE duels
			SET status = 'finished',
			    winner_user_id = $2,
			    finished_at = NOW()
			WHERE id = $1
		`, duelID, winnerID)
		if err != nil {
			log.Println("Ошибка завершения дуэли:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("Ошибка сохранения дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeDuelJSON(w, http.StatusOK, map[string]any{
		"correct":  correct,
		"finished": finished,
	})
}
