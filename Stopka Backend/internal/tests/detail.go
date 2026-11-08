
package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetFinishedAttempt(
	ctx context.Context,
	attemptID int64,
	userID int64,
) (*HistoryItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var item HistoryItem
	var deckTitle string

	err := r.db.QueryRow(ctx, `
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
		WHERE ta.id = $1
		  AND ta.user_id = $2
		  AND ta.status = 'finished'
	`, attemptID, userID).Scan(
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

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получить результат теста: %w", err)
	}

	item.Deck = HistoryDeck{
		ID:    item.DeckID,
		Title: deckTitle,
	}

	return &item, nil
}
