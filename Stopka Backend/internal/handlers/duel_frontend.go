
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type duelClientPlayer struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
	Answered int    `json:"answered"`
	IsMe     bool   `json:"is_me"`
}

type duelClientDeck struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type duelClientItem struct {
	ID            int64              `json:"id"`
	Code          string             `json:"code"`
	Status        string             `json:"status"`
	Deck          duelClientDeck     `json:"deck"`
	QuestionCount int                `json:"question_count"`
	Players       []duelClientPlayer `json:"players"`
	Result        string             `json:"result"`
}

type duelClientOption struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type duelClientQuestion struct {
	ID      int64              `json:"id"`
	Text    string             `json:"text"`
	Options []duelClientOption `json:"options"`
}

type duelClientAnswer struct {
	QuestionID int64 `json:"question_id"`
	OptionID   int64 `json:"option_id"`
	Correct    bool  `json:"correct"`
}

type duelClientGame struct {
	duelClientItem
	Questions []duelClientQuestion `json:"questions"`
	MyAnswers []duelClientAnswer   `json:"my_answers"`
}

func duelResult(status string, winner *int64, userID int64) string {
	if status != "finished" {
		return ""
	}
	if winner == nil {
		return "draw"
	}
	if *winner == userID {
		return "win"
	}
	return "lose"
}

func duelUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := AuthenticatedUserID(r.Context())
	if !ok || id <= 0 {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return 0, false
	}
	return id, true
}

func duelPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Неверный ID дуэли", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func (h *DuelsHandler) ListFrontend(w http.ResponseWriter, r *http.Request) {
	userID, ok := duelUser(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, `
		SELECT d.id, d.code, d.status, de.id, de.title,
		       d.winner_user_id,
		       (SELECT COUNT(*)::int
		        FROM duel_questions dq
		        WHERE dq.duel_id=d.id),
		       dp.user_id,
		       COALESCE(
		           NULLIF(TRIM(CONCAT(u.first_name,' ',u.last_name)), ''),
		           'Ученик'
		       ),
		       dp.score, dp.answered
		FROM duels d
		JOIN decks de ON de.id=d.deck_id
		LEFT JOIN duel_players dp ON dp.duel_id=d.id
		LEFT JOIN users u ON u.id=dp.user_id
		WHERE d.status='waiting'
		   OR EXISTS (
		       SELECT 1 FROM duel_players mine
		       WHERE mine.duel_id=d.id AND mine.user_id=$1
		   )
		ORDER BY d.id DESC, dp.joined_at, dp.user_id
	`, userID)
	if err != nil {
		apiFail(w, err)
		return
	}

	all := make([]duelClientItem, 0)
	winners := make(map[int64]*int64)
	positions := make(map[int64]int)

	for rows.Next() {
		var item duelClientItem
		var winner, playerID *int64
		var name string
		var score, answered *int

		err = rows.Scan(
			&item.ID,
			&item.Code,
			&item.Status,
			&item.Deck.ID,
			&item.Deck.Title,
			&winner,
			&item.QuestionCount,
			&playerID,
			&name,
			&score,
			&answered,
		)
		if err != nil {
			break
		}

		index, exists := positions[item.ID]
		if !exists {
			index = len(all)
			positions[item.ID] = index
			item.Players = make([]duelClientPlayer, 0, 2)
			all = append(all, item)
			winners[item.ID] = winner
		}

		if playerID != nil {
			player := duelClientPlayer{
				UserID: *playerID,
				Name:   name,
				IsMe:   *playerID == userID,
			}
			if score != nil {
				player.Score = *score
			}
			if answered != nil {
				player.Answered = *answered
			}
			all[index].Players = append(all[index].Players, player)
		}
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	open := make([]duelClientItem, 0)
	mine := make([]duelClientItem, 0)
	history := make([]duelClientItem, 0)

	for _, d := range all {
		member := false

		for _, p := range d.Players {
			if p.IsMe {
				member = true
				break
			}
		}

		if member {
			d.Result = duelResult(d.Status, winners[d.ID], userID)

			if d.Status == "waiting" || d.Status == "active" {
				mine = append(mine, d)
			} else {
				history = append(history, d)
			}
		} else if d.Status == "waiting" {
			open = append(open, d)
		}
	}

	sendJSON(w, map[string]any{
		"open":    open,
		"mine":    mine,
		"history": history,
	})
}

func (h *DuelsHandler) loadFrontendGame(
	ctx context.Context,
	userID, duelID int64,
) (duelClientGame, error) {
	var game duelClientGame

	game.Questions = make([]duelClientQuestion, 0)
	game.MyAnswers = make([]duelClientAnswer, 0)
	game.Players = make([]duelClientPlayer, 0, 2)

	var winner *int64

	err := h.db.QueryRow(ctx, `
		SELECT d.id, d.code, d.status,
		       de.id, de.title, d.winner_user_id,
		       (SELECT COUNT(*)::int
		        FROM duel_questions dq
		        WHERE dq.duel_id=d.id)
		FROM duels d
		JOIN decks de ON de.id=d.deck_id
		WHERE d.id=$1
		  AND EXISTS (
		      SELECT 1 FROM duel_players dp
		      WHERE dp.duel_id=d.id AND dp.user_id=$2
		  )
	`, duelID, userID).Scan(
		&game.ID,
		&game.Code,
		&game.Status,
		&game.Deck.ID,
		&game.Deck.Title,
		&winner,
		&game.QuestionCount,
	)
	if err != nil {
		return game, err
	}

	game.Result = duelResult(game.Status, winner, userID)

	players, err := h.db.Query(ctx, `
		SELECT dp.user_id,
		       COALESCE(
		           NULLIF(TRIM(CONCAT(u.first_name,' ',u.last_name)), ''),
		           'Ученик'
		       ),
		       dp.score, dp.answered
		FROM duel_players dp
		JOIN users u ON u.id=dp.user_id
		WHERE dp.duel_id=$1
		ORDER BY dp.joined_at, dp.user_id
	`, duelID)
	if err != nil {
		return game, err
	}

	for players.Next() {
		var p duelClientPlayer

		err = players.Scan(
			&p.UserID,
			&p.Name,
			&p.Score,
			&p.Answered,
		)
		if err != nil {
			break
		}

		p.IsMe = p.UserID == userID
		game.Players = append(game.Players, p)
	}

	if err == nil {
		err = players.Err()
	}
	players.Close()

	if err != nil {
		return game, err
	}

	rows, err := h.db.Query(ctx, `
		SELECT q.id, q.text, qo.id, qo.text
		FROM duel_questions dq
		JOIN questions q ON q.id=dq.question_id
		LEFT JOIN question_options qo ON qo.question_id=q.id
		WHERE dq.duel_id=$1
		ORDER BY dq.position, q.id, qo.position, qo.id
	`, duelID)
	if err != nil {
		return game, err
	}

	positions := make(map[int64]int)

	for rows.Next() {
		var qID int64
		var qText string
		var oID *int64
		var oText *string

		err = rows.Scan(&qID, &qText, &oID, &oText)
		if err != nil {
			break
		}

		index, exists := positions[qID]
		if !exists {
			index = len(game.Questions)
			positions[qID] = index

			game.Questions = append(
				game.Questions,
				duelClientQuestion{
					ID:      qID,
					Text:    qText,
					Options: make([]duelClientOption, 0),
				},
			)
		}

		if oID != nil && oText != nil {
			game.Questions[index].Options = append(
				game.Questions[index].Options,
				duelClientOption{
					ID:   *oID,
					Text: *oText,
				},
			)
		}
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		return game, err
	}

	answers, err := h.db.Query(ctx, `
		SELECT question_id, option_id, is_correct
		FROM duel_answers
		WHERE duel_id=$1 AND user_id=$2
	`, duelID, userID)
	if err != nil {
		return game, err
	}

	for answers.Next() {
		var a duelClientAnswer
		err = answers.Scan(
			&a.QuestionID,
			&a.OptionID,
			&a.Correct,
		)
		if err != nil {
			break
		}
		game.MyAnswers = append(game.MyAnswers, a)
	}

	if err == nil {
		err = answers.Err()
	}
	answers.Close()

	return game, err
}

func (h *DuelsHandler) GetFrontendGame(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := duelUser(w, r)
	if !ok {
		return
	}

	duelID, ok := duelPathID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	game, err := h.loadFrontendGame(ctx, userID, duelID)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Дуэль не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{"duel": game})
}

func (h *DuelsHandler) AnswerFrontend(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := duelUser(w, r)
	if !ok {
		return
	}

	duelID, ok := duelPathID(w, r)
	if !ok {
		return
	}

	var req struct {
		QuestionID int64 `json:"question_id"`
		OptionID   int64 `json:"option_id"`
	}

	if err := json.NewDecoder(
		http.MaxBytesReader(w, r.Body, 4096),
	).Decode(&req); err != nil ||
		req.QuestionID <= 0 || req.OptionID <= 0 {
		http.Error(w, "Неверный ответ", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		apiFail(w, err)
		return
	}
	defer tx.Rollback(ctx)

	var status string

	err = tx.QueryRow(ctx, `
		SELECT status
		FROM duels
		WHERE id=$1
		FOR UPDATE
	`, duelID).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Дуэль не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		apiFail(w, err)
		return
	}

	if status != "active" {
		http.Error(w, "Дуэль не активна", http.StatusConflict)
		return
	}

	var isPlayer bool

	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM duel_players
			WHERE duel_id=$1 AND user_id=$2
		)
	`, duelID, userID).Scan(&isPlayer)

	if err != nil {
		apiFail(w, err)
		return
	}
	if !isPlayer {
		http.Error(w, "Нет доступа", http.StatusForbidden)
		return
	}

	var correct bool
	var correctOptionID int64

	err = tx.QueryRow(ctx, `
		SELECT chosen.is_correct, right_option.id
		FROM duel_questions dq
		JOIN question_options chosen
		     ON chosen.question_id=dq.question_id
		JOIN question_options right_option
		     ON right_option.question_id=dq.question_id
		    AND right_option.is_correct
		WHERE dq.duel_id=$1
		  AND dq.question_id=$2
		  AND chosen.id=$3
	`, duelID, req.QuestionID, req.OptionID).Scan(
		&correct,
		&correctOptionID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(
			w,
			"Вопрос или ответ не найден",
			http.StatusBadRequest,
		)
		return
	}
	if err != nil {
		apiFail(w, err)
		return
	}

	var inserted int64

	err = tx.QueryRow(ctx, `
		INSERT INTO duel_answers(
			duel_id,
			user_id,
			question_id,
			option_id,
			is_correct
		)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (duel_id,user_id,question_id)
		DO NOTHING
		RETURNING question_id
	`, duelID, userID, req.QuestionID, req.OptionID, correct).Scan(
		&inserted,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "На вопрос уже ответили", http.StatusConflict)
		return
	}
	if err != nil {
		apiFail(w, err)
		return
	}

	increment := 0
	if correct {
		increment = 1
	}

	_, err = tx.Exec(ctx, `
		UPDATE duel_players
		SET score=score+$3,
		    answered=answered+1
		WHERE duel_id=$1 AND user_id=$2
	`, duelID, userID, increment)

	if err != nil {
		apiFail(w, err)
		return
	}

	var total, completed int

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM duel_questions
		WHERE duel_id=$1
	`, duelID).Scan(&total)

	if err != nil {
		apiFail(w, err)
		return
	}

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM duel_players
		WHERE duel_id=$1 AND answered >= $2
	`, duelID, total).Scan(&completed)

	if err != nil {
		apiFail(w, err)
		return
	}

	if total > 0 && completed == 2 {
		var winner *int64

		err = tx.QueryRow(ctx, `
			SELECT CASE
				WHEN MAX(score)=MIN(score) THEN NULL
				ELSE (
					SELECT user_id
					FROM duel_players
					WHERE duel_id=$1
					ORDER BY score DESC,user_id
					LIMIT 1
				)
			END
			FROM duel_players
			WHERE duel_id=$1
		`, duelID).Scan(&winner)

		if err != nil {
			apiFail(w, err)
			return
		}

		_, err = tx.Exec(ctx, `
			UPDATE duels
			SET status='finished',
			    winner_user_id=$2,
			    finished_at=NOW()
			WHERE id=$1
		`, duelID, winner)

		if err != nil {
			apiFail(w, err)
			return
		}
	}

	if err = tx.Commit(ctx); err != nil {
		apiFail(w, err)
		return
	}

	game, err := h.loadFrontendGame(ctx, userID, duelID)
	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{
		"correct_option_id": correctOptionID,
		"correct":           correct,
		"duel":              game,
	})
}

func (h *DuelsHandler) LeaveFrontend(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := duelUser(w, r)
	if !ok {
		return
	}

	duelID, ok := duelPathID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		apiFail(w, err)
		return
	}
	defer tx.Rollback(ctx)

	var status string

	err = tx.QueryRow(ctx, `
		SELECT status
		FROM duels
		WHERE id=$1
		FOR UPDATE
	`, duelID).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Дуэль не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		apiFail(w, err)
		return
	}

	var isPlayer bool

	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM duel_players
			WHERE duel_id=$1 AND user_id=$2
		)
	`, duelID, userID).Scan(&isPlayer)

	if err != nil {
		apiFail(w, err)
		return
	}
	if !isPlayer {
		http.Error(w, "Нет доступа", http.StatusForbidden)
		return
	}

	if status == "waiting" {
		_, err = tx.Exec(ctx, `
			UPDATE duels
			SET status='cancelled',
			    finished_at=NOW()
			WHERE id=$1
		`, duelID)
	} else if status == "active" {
		var opponentID int64

		err = tx.QueryRow(ctx, `
			SELECT user_id
			FROM duel_players
			WHERE duel_id=$1 AND user_id<>$2
			LIMIT 1
		`, duelID, userID).Scan(&opponentID)

		if err == nil {
			_, err = tx.Exec(ctx, `
				UPDATE duels
				SET status='finished',
				    winner_user_id=$2,
				    finished_at=NOW()
				WHERE id=$1
			`, duelID, opponentID)
		}
	} else {
		http.Error(w, "Дуэль уже завершена", http.StatusConflict)
		return
	}

	if err != nil {
		apiFail(w, err)
		return
	}

	if err = tx.Commit(ctx); err != nil {
		apiFail(w, err)
		return
	}

	game, err := h.loadFrontendGame(ctx, userID, duelID)
	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{"duel": game})
}
