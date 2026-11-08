
package tests

import (
	"context"
	"fmt"
	"time"
)

type ReviewOption struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type ReviewQuestion struct {
	Text            string         `json:"text"`
	Correct         bool           `json:"correct"`
	Options         []ReviewOption `json:"options"`
	ChosenOptionID  int64          `json:"chosen_option_id"`
	CorrectOptionID int64          `json:"correct_option_id"`
}

func (r *Repository) GetAttemptReview(
	ctx context.Context,
	attemptID int64,
	userID int64,
) ([]ReviewQuestion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()


	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM test_attempts
			WHERE id = $1
			  AND user_id = $2
			  AND status = 'finished'
		)
	`, attemptID, userID).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("проверить попытку: %w", err)
	}

	if !exists {
		return nil, ErrAttemptNotFound
	}


	rows, err := r.db.Query(ctx, `
		SELECT
			q.id,
			q.text,
			ta.option_id,
			ta.is_correct,
			qo.id,
			qo.text,
			qo.is_correct
		FROM test_attempt_questions taq
		JOIN questions q
			ON q.id = taq.question_id
		JOIN test_answers ta
			ON ta.attempt_id = taq.attempt_id
			AND ta.question_id = taq.question_id
		JOIN question_options qo
			ON qo.question_id = q.id
		WHERE taq.attempt_id = $1
		ORDER BY
			taq.position,
			q.id,
			qo.position,
			qo.id
	`, attemptID)

	if err != nil {
		return nil, fmt.Errorf("получить разбор теста: %w", err)
	}
	defer rows.Close()

	review := make([]ReviewQuestion, 0)
	questionIndex := make(map[int64]int)

	for rows.Next() {
		var questionID int64
		var questionText string
		var chosenOptionID int64
		var correct bool

		var optionID int64
		var optionText string
		var optionIsCorrect bool

		if err := rows.Scan(
			&questionID,
			&questionText,
			&chosenOptionID,
			&correct,
			&optionID,
			&optionText,
			&optionIsCorrect,
		); err != nil {
			return nil, fmt.Errorf("прочитать разбор: %w", err)
		}

		index, found := questionIndex[questionID]

		if !found {
			index = len(review)
			questionIndex[questionID] = index

			review = append(review, ReviewQuestion{
				Text:           questionText,
				Correct:        correct,
				ChosenOptionID: chosenOptionID,
				Options:        make([]ReviewOption, 0),
			})
		}

		review[index].Options = append(
			review[index].Options,
			ReviewOption{
				ID:   optionID,
				Text: optionText,
			},
		)

		if optionIsCorrect {
			review[index].CorrectOptionID = optionID
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения разбора: %w", err)
	}

	return review, nil
}
