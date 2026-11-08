package tests

import (
	"context"
	"fmt"
)

func (r *Repository) GetClassDeckResults(
	ctx context.Context,
	classID int64,
	deckID int64,
	teacherIDs ...int64,
) ([]TestResult, error) {
	if len(teacherIDs) != 1 || teacherIDs[0] <= 0 {
		return nil, ErrStatisticsForbidden
	}

	teacherID := teacherIDs[0]

	if err := r.checkClassAccess(ctx, classID, teacherID); err != nil {
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
		JOIN class_members cm ON cm.user_id = u.id
		JOIN classes c ON c.id = cm.class_id
		WHERE cm.class_id = $1
		  AND c.teacher_id = $3
		  AND ta.deck_id = $2
		  AND ta.status = 'finished'
		ORDER BY ta.finished_at DESC, ta.id DESC
	`, classID, deckID, teacherID)

	if err != nil {
		return nil, fmt.Errorf("получить результаты класса: %w", err)
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
			return nil, fmt.Errorf("прочитать результат ученика: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список результатов: %w", err)
	}

	return results, nil
}
