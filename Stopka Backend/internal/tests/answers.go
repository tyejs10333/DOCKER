package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrAttemptNotFound = errors.New("попытка не найдена или уже завершена")
	ErrInvalidAnswer   = errors.New("вопрос или вариант ответа не подходит")
	ErrAlreadyAnswered = errors.New("на этот вопрос уже ответили")
)

func (r *Repository) SaveAnswer(
	ctx context.Context,
	attemptID int64,
	userID int64,
	questionID int64,
	optionID int64,
) (*Answer, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string

	err = tx.QueryRow(ctx, `
		SELECT status
		FROM test_attempts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, attemptID, userID).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить попытку: %w", err)
	}
	if status != "in_progress" {
		return nil, ErrAttemptNotFound
	}

	var isCorrect bool

	err = tx.QueryRow(ctx, `
		SELECT qo.is_correct
		FROM test_attempt_questions taq
		JOIN question_options qo
			ON qo.question_id = taq.question_id
		WHERE taq.attempt_id = $1
		  AND taq.question_id = $2
		  AND qo.id = $3
	`, attemptID, questionID, optionID).Scan(&isCorrect)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidAnswer
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось проверить ответ: %w", err)
	}

	var alreadyAnswered bool

	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM test_answers
			WHERE attempt_id = $1 AND question_id = $2
		)
	`, attemptID, questionID).Scan(&alreadyAnswered)
	if err != nil {
		return nil, fmt.Errorf("не удалось проверить повторный ответ: %w", err)
	}

	if alreadyAnswered {
		return nil, ErrAlreadyAnswered
	}

	answer := &Answer{}

	err = tx.QueryRow(ctx, `
		INSERT INTO test_answers (
			attempt_id,
			question_id,
			option_id,
			is_correct
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			attempt_id,
			question_id,
			option_id,
			is_correct,
			answered_at
	`, attemptID, questionID, optionID, isCorrect).Scan(
		&answer.AttemptID,
		&answer.QuestionID,
		&answer.OptionID,
		&answer.IsCorrect,
		&answer.AnsweredAt,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить ответ: %w", err)
	}

	if isCorrect {
		_, err = tx.Exec(ctx, `
			UPDATE test_attempts
			SET correct = correct + 1
			WHERE id = $1
		`, attemptID)
		if err != nil {
			return nil, fmt.Errorf("не удалось обновить результат: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("не удалось завершить транзакцию: %w", err)
	}

	return answer, nil
}


func (r *Repository) AnswerFeedback(ctx context.Context, attemptID, userID, questionID int64) (int64, int, int, error) {
	var correctOptionID int64
	var correct, total int
	err := r.db.QueryRow(ctx, `
 SELECT qo.id, ta.correct, (SELECT COUNT(*)::int FROM test_answers WHERE attempt_id=$1)
 FROM test_attempts ta
 JOIN test_attempt_questions taq ON taq.attempt_id=ta.id AND taq.question_id=$3
 JOIN question_options qo ON qo.question_id=taq.question_id AND qo.is_correct=TRUE
 WHERE ta.id=$1 AND ta.user_id=$2
 `, attemptID, userID, questionID).Scan(&correctOptionID, &correct, &total)
	if err != nil {
		return 0, 0, 0, err
	}
	return correctOptionID, correct, total - correct, nil
}
