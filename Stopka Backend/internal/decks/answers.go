
package decks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) CheckAnswer(
	ctx context.Context,
	questionID int64,
	optionID int64,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var isCorrect bool

	err := r.db.QueryRow(ctx, `
		SELECT is_correct
		FROM question_options
		WHERE id = $1 AND question_id = $2
	`, optionID, questionID).Scan(&isCorrect)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("вариант ответа не найден")
	}

	if err != nil {
		return false, fmt.Errorf("ошибка проверки ответа: %w", err)
	}

	return isCorrect, nil
}
