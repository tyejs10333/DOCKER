
package decks

import (
	"context"
	"fmt"
	"time"
)

type Question struct {
	ID       int64            `json:"id"`
	DeckID   int64            `json:"deck_id"`
	Text     string           `json:"text"`
	Position int              `json:"position"`
	Options  []QuestionOption `json:"options"`
}

type QuestionOption struct {
	ID       int64  `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

func (r *Repository) GetQuestions(ctx context.Context, deckID int64) ([]Question, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT
			q.id,
			q.deck_id,
			q.text,
			q.position,
			o.id,
			o.text,
			o.position
		FROM questions q
		LEFT JOIN question_options o ON o.question_id = q.id
		WHERE q.deck_id = $1
		ORDER BY q.position, q.id, o.position, o.id
	`, deckID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить вопросы: %w", err)
	}
	defer rows.Close()

	questions := make([]Question, 0)
	questionIndex := make(map[int64]int)

	for rows.Next() {
		var question Question
		var optionID *int64
		var optionText *string
		var optionPosition *int

		if err := rows.Scan(
			&question.ID,
			&question.DeckID,
			&question.Text,
			&question.Position,
			&optionID,
			&optionText,
			&optionPosition,
		); err != nil {
			return nil, fmt.Errorf("не удалось прочитать вопрос: %w", err)
		}

		index, exists := questionIndex[question.ID]
		if !exists {
			question.Options = make([]QuestionOption, 0)
			index = len(questions)
			questionIndex[question.ID] = index
			questions = append(questions, question)
		}

		if optionID != nil && optionText != nil && optionPosition != nil {
			questions[index].Options = append(
				questions[index].Options,
				QuestionOption{
					ID:       *optionID,
					Text:     *optionText,
					Position: *optionPosition,
				},
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения вопросов: %w", err)
	}

	return questions, nil
}
