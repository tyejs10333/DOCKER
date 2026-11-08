
package decks

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)


type Deck struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Topic         string `json:"topic"`
	Period        string `json:"period"`
	Position      int    `json:"position"`
	CardCount     int    `json:"card_count"`
	QuestionCount int    `json:"question_count"`
}


type Repository struct {
	db *pgxpool.Pool
}


func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}


func (r *Repository) GetAll(ctx context.Context) ([]Deck, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT
			d.id,
			d.title,
			d.topic,
			d.period,
			d.position,
			(SELECT COUNT(*) FROM cards WHERE deck_id = d.id) AS card_count,
			(SELECT COUNT(*) FROM questions WHERE deck_id = d.id) AS question_count
		FROM decks d
		ORDER BY d.position, d.id
	`)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить колоды: %w", err)
	}
	defer rows.Close()

	decks := make([]Deck, 0)

	for rows.Next() {
		var deck Deck

		if err := rows.Scan(
			&deck.ID,
			&deck.Title,
			&deck.Topic,
			&deck.Period,
			&deck.Position,
			&deck.CardCount,
			&deck.QuestionCount,
		); err != nil {
			return nil, fmt.Errorf("не удалось прочитать колоду: %w", err)
		}

		decks = append(decks, deck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения колод: %w", err)
	}

	return decks, nil
}
