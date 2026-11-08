
package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrAttemptIncomplete = errors.New("не на все вопросы даны ответы")
	ErrAttemptFinished   = errors.New("тест уже завершён")
)

func (r *Repository) FinishAttempt(
	ctx context.Context,
	attemptID int64,
	userID int64,
) (*Attempt, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	var attempt Attempt

	err = tx.QueryRow(ctx, `
		SELECT
			id, user_id, deck_id, status,
			correct, total, percent,
			started_at, finished_at
		FROM test_attempts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, attemptID, userID).Scan(
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

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получить попытку: %w", err)
	}

	if attempt.Status != "in_progress" {
		return nil, ErrAttemptFinished
	}

	var answeredCount int

	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM test_answers
		WHERE attempt_id = $1
	`, attemptID).Scan(&answeredCount)

	if err != nil {
		return nil, fmt.Errorf("посчитать ответы: %w", err)
	}

	if answeredCount != attempt.Total {
		return nil, ErrAttemptIncomplete
	}

	if attempt.Total <= 0 {
		return nil, fmt.Errorf("в тесте нет вопросов")
	}

	percent := attempt.Correct * 100 / attempt.Total

	var finishedAt time.Time

	err = tx.QueryRow(ctx, `
		UPDATE test_attempts
		SET
			status = 'finished',
			percent = $2,
			finished_at = NOW()
		WHERE id = $1
		RETURNING finished_at
	`, attemptID, percent).Scan(&finishedAt)

	if err != nil {
		return nil, fmt.Errorf("завершить попытку: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("сохранить завершение теста: %w", err)
	}

	attempt.Status = "finished"
	attempt.Percent = percent
	attempt.FinishedAt = &finishedAt

	return &attempt, nil
}
