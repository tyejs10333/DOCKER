package tests

import (
	"context"
	"fmt"
)

type QuestionStatistics struct {
	QuestionID   int64   `json:"question_id"`
	QuestionText string  `json:"question_text"`
	AnswersCount int     `json:"answers_count"`
	CorrectCount int     `json:"correct_count"`
	WrongCount   int     `json:"wrong_count"`
	ErrorPercent float64 `json:"error_percent"`
}

func (r *Repository) GetQuestionStatistics(
	ctx context.Context,
	deckID int64,
	teacherIDs ...int64,
) ([]QuestionStatistics, error) {
	if len(teacherIDs) != 1 || teacherIDs[0] <= 0 {
		return nil, ErrStatisticsForbidden
	}

	teacherID := teacherIDs[0]

	if err := r.checkTeacherAccess(ctx, teacherID); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			q.id,
			q.text,
			COUNT(ta.id)::integer AS answers_count,
			COUNT(ta.id) FILTER (
				WHERE ans.is_correct = true
			)::integer AS correct_count,
			COUNT(ta.id) FILTER (
				WHERE ans.is_correct = false
			)::integer AS wrong_count,
			COALESCE(
				ROUND(
					100.0 * COUNT(ta.id) FILTER (
						WHERE ans.is_correct = false
					) / NULLIF(COUNT(ta.id), 0),
					1
				),
				0
			)::double precision AS error_percent
		FROM questions q
		LEFT JOIN test_answers ans
			ON ans.question_id = q.id
		LEFT JOIN test_attempts ta
			ON ta.id = ans.attempt_id
			AND ta.status = 'finished'
			AND EXISTS (
				SELECT 1
				FROM class_members cm
				JOIN classes c ON c.id = cm.class_id
				WHERE cm.user_id = ta.user_id
				  AND c.teacher_id = $2
			)
		WHERE q.deck_id = $1
		GROUP BY q.id, q.text
		ORDER BY error_percent DESC, q.id
	`, deckID, teacherID)

	if err != nil {
		return nil, fmt.Errorf("получить статистику вопросов: %w", err)
	}
	defer rows.Close()

	result := make([]QuestionStatistics, 0)

	for rows.Next() {
		var item QuestionStatistics

		if err := rows.Scan(
			&item.QuestionID,
			&item.QuestionText,
			&item.AnswersCount,
			&item.CorrectCount,
			&item.WrongCount,
			&item.ErrorPercent,
		); err != nil {
			return nil, fmt.Errorf("прочитать статистику вопроса: %w", err)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список вопросов: %w", err)
	}

	return result, nil
}
