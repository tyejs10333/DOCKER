
package decks

import (
	"context"
	"fmt"
	"time"
)

type Card struct {
	ID       int64  `json:"id"`
	DeckID   int64  `json:"deck_id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Position int    `json:"position"`
}

func (r *Repository) GetCards(ctx context.Context, deckID int64) ([]Card, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT id, deck_id, question, answer, position
		FROM cards
		WHERE deck_id = $1
		ORDER BY position, id
	`, deckID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карточки: %w", err)
	}
	defer rows.Close()

	cards := make([]Card, 0)

	for rows.Next() {
		var card Card

		if err := rows.Scan(
			&card.ID,
			&card.DeckID,
			&card.Question,
			&card.Answer,
			&card.Position,
		); err != nil {
			return nil, fmt.Errorf("не удалось прочитать карточку: %w", err)
		}

		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения карточек: %w", err)
	}

	return cards, nil
}
