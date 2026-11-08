
package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAttempt(
	ctx context.Context,
	userID int64,
	deckID int64,
) (*Attempt, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	var total int

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM questions
		WHERE deck_id = $1
	`, deckID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("не удалось подсчитать вопросы: %w", err)
	}

	if total == 0 {
		return nil, fmt.Errorf("в выбранной колоде нет вопросов")
	}

	attempt := &Attempt{}

	err = tx.QueryRow(ctx, `
		INSERT INTO test_attempts (user_id, deck_id, total)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			user_id,
			deck_id,
			status,
			correct,
			total,
			percent,
			started_at,
			finished_at
	`, userID, deckID, total).Scan(
		&attempt.ID,
		&attempt.UserID,
		&attempt.DeckID,
		&attempt.Status,
		&attempt.Correct,
		&attempt.Total,
		&attempt.Percent,
		&attempt.StartedAt,
		&attempt.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать попытку: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO test_attempt_questions (
			attempt_id,
			question_id,
			position
		)
		SELECT
			$1,
			id,
			position
		FROM questions
		WHERE deck_id = $2
	`, attempt.ID, deckID)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить вопросы попытки: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("не удалось сохранить попытку: %w", err)
	}

	return attempt, nil
}
