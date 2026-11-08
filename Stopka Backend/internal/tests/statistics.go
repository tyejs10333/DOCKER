package tests

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrStatisticsForbidden = errors.New("нет доступа к статистике")

type TestResult struct {
	AttemptID  int64     `json:"attempt_id"`
	UserID     int64     `json:"user_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Correct    int       `json:"correct"`
	Total      int       `json:"total"`
	Percent    int       `json:"percent"`
	FinishedAt time.Time `json:"finished_at"`
}

func (r *Repository) checkTeacherAccess(
	ctx context.Context,
	teacherID int64,
) error {
	var allowed bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1 AND role = 'teacher' AND teacher_verified = TRUE
		)
	`, teacherID).Scan(&allowed)

	if err != nil {
		return fmt.Errorf("проверить роль учителя: %w", err)
	}
	if !allowed {
		return ErrStatisticsForbidden
	}
	return nil
}

func (r *Repository) checkClassAccess(
	ctx context.Context,
	classID int64,
	teacherID int64,
) error {
	var allowed bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM classes c
			JOIN users u ON u.id = c.teacher_id
			WHERE c.id = $1
			  AND c.teacher_id = $2
			  AND u.role = 'teacher' AND u.teacher_verified = TRUE
		)
	`, classID, teacherID).Scan(&allowed)

	if err != nil {
		return fmt.Errorf("проверить доступ к классу: %w", err)
	}
	if !allowed {
		return ErrStatisticsForbidden
	}
	return nil
}

func (r *Repository) GetDeckResults(
	ctx context.Context,
	deckID int64,
	teacherIDs ...int64,
) ([]TestResult, error) {
	if len(teacherIDs) != 1 || teacherIDs[0] <= 0 {
		return nil, ErrStatisticsForbidden
	}

	teacherID := teacherIDs[0]

	if err := r.checkTeacherAccess(ctx, teacherID); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			ta.id,
			u.id,
			u.first_name,
			u.last_name,
			ta.correct,
			ta.total,
			ta.percent,
			ta.finished_at
		FROM test_attempts ta
		JOIN users u ON u.id = ta.user_id
		WHERE ta.deck_id = $1
		  AND ta.status = 'finished'
		  AND EXISTS (
			SELECT 1
			FROM class_members cm
			JOIN classes c ON c.id = cm.class_id
			WHERE cm.user_id = ta.user_id
			  AND c.teacher_id = $2
		  )
		ORDER BY ta.finished_at DESC, ta.id DESC
	`, deckID, teacherID)

	if err != nil {
		return nil, fmt.Errorf("получить результаты тестов: %w", err)
	}
	defer rows.Close()

	results := make([]TestResult, 0)

	for rows.Next() {
		var result TestResult

		if err := rows.Scan(
			&result.AttemptID,
			&result.UserID,
			&result.FirstName,
			&result.LastName,
			&result.Correct,
			&result.Total,
			&result.Percent,
			&result.FinishedAt,
		); err != nil {
			return nil, fmt.Errorf("прочитать результат: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать результаты: %w", err)
	}

	return results, nil
}
