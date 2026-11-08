
package decks

import (
	"context"
	"fmt"
	"time"
)

type CardProgress struct {
	Studied       int        `json:"studied"`
	Known         int        `json:"known"`
	Weak          int        `json:"weak"`
	LastStudiedAt *time.Time `json:"last_studied_at"`
}

type DeckWithProgress struct {
	Deck
	Progress CardProgress `json:"progress"`
}

func (r *Repository) GetAllForUser(ctx context.Context, userID int64) ([]DeckWithProgress, error) {
	deckList, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			c.deck_id,
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE cr.grade = 'good')::int,
			COUNT(*) FILTER (WHERE cr.grade IN ('again', 'hard'))::int,
			MAX(cr.reviewed_at)
		FROM card_reviews cr
		JOIN cards c ON c.id = cr.card_id
		WHERE cr.user_id = $1
		GROUP BY c.deck_id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения прогресса: %w", err)
	}
	defer rows.Close()

	progress := make(map[int64]CardProgress)

	for rows.Next() {
		var deckID int64
		var p CardProgress

		if err := rows.Scan(
			&deckID,
			&p.Studied,
			&p.Known,
			&p.Weak,
			&p.LastStudiedAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка чтения прогресса: %w", err)
		}

		progress[deckID] = p
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]DeckWithProgress, 0, len(deckList))
	for _, deck := range deckList {
		result = append(result, DeckWithProgress{
			Deck:     deck,
			Progress: progress[deck.ID],
		})
	}

	return result, nil
}

type StudyCard struct {
	ID       int64  `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Level    int    `json:"level"`
}

func (r *Repository) GetStudyCards(
	ctx context.Context,
	deckID int64,
	userID int64,
	mode string,
) ([]StudyCard, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT
			c.id,
			c.question,
			c.answer,
			CASE
				WHEN cr.grade = 'good' THEN 2
				WHEN cr.grade = 'hard' THEN 1
				ELSE 0
			END AS level
		FROM cards c
		LEFT JOIN card_reviews cr
			ON cr.card_id = c.id
			AND cr.user_id = $2
		WHERE c.deck_id = $1
			AND (
				$3 = 'all'
				OR ($3 = 'new' AND cr.card_id IS NULL)
				OR ($3 = 'weak' AND cr.grade IN ('again', 'hard'))
			)
		ORDER BY c.position, c.id
	`, deckID, userID, mode)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения учебных карточек: %w", err)
	}
	defer rows.Close()

	result := make([]StudyCard, 0)

	for rows.Next() {
		var card StudyCard
		if err := rows.Scan(
			&card.ID,
			&card.Question,
			&card.Answer,
			&card.Level,
		); err != nil {
			return nil, fmt.Errorf("ошибка чтения карточки: %w", err)
		}
		result = append(result, card)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
