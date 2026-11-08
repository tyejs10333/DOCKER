
package tests

import (
	"context"
	"fmt"
	"time"
)

type HistoryItem struct {
	ID         int64      `json:"id"`
	DeckID     int64      `json:"deck_id"`
	Deck       HistoryDeck `json:"deck"`
	Status     string     `json:"status"`
	Correct    int        `json:"correct"`
	Total      int        `json:"total"`
	Percent    int        `json:"percent"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

type HistoryDeck struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func (r *Repository) GetUserTestHistory(
	ctx context.Context,
	userID int64,
) ([]HistoryItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT
			ta.id,
			ta.deck_id,
			d.title,
			ta.status,
			ta.correct,
			ta.total,
			ta.percent,
			ta.started_at,
			ta.finished_at
		FROM test_attempts ta
		JOIN decks d ON d.id = ta.deck_id
		WHERE ta.user_id = $1
		  AND ta.status = 'finished'
		ORDER BY ta.finished_at DESC, ta.id DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("получить историю тестов: %w", err)
	}
	defer rows.Close()

	history := make([]HistoryItem, 0)

	for rows.Next() {
		var item HistoryItem
		var deckTitle string

		err := rows.Scan(
			&item.ID,
			&item.DeckID,
			&deckTitle,
			&item.Status,
			&item.Correct,
			&item.Total,
			&item.Percent,
			&item.StartedAt,
			&item.FinishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("прочитать историю тестов: %w", err)
		}

		item.Deck = HistoryDeck{
			ID:    item.DeckID,
			Title: deckTitle,
		}

		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения истории: %w", err)
	}

	return history, nil
}
