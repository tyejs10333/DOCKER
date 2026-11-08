package classes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var (
	ErrClassNotFound = errors.New("класс не найден")
	ErrAlreadyMember = errors.New("ученик уже состоит в классе")
	ErrTeacherJoin   = errors.New("учитель не может вступить в свой класс как ученик")
	ErrNotStudent    = errors.New("вступать в классы могут только ученики")
)

func (r *Repository) JoinClass(
	ctx context.Context,
	userID int64,
	code string,
) (*Class, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	var role string

	err = tx.QueryRow(ctx, `
		SELECT role
		FROM users
		WHERE id = $1
	`, userID).Scan(&role)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotStudent
	}
	if err != nil {
		return nil, fmt.Errorf("проверить роль пользователя: %w", err)
	}

	if role != "student" {
		return nil, ErrNotStudent
	}

	var class Class

	err = tx.QueryRow(ctx, `
		SELECT id, teacher_id, name, code, created_at
		FROM classes
		WHERE code = $1
	`, code).Scan(
		&class.ID,
		&class.TeacherID,
		&class.Name,
		&class.Code,
		&class.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrClassNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("найти класс: %w", err)
	}

	if class.TeacherID == userID {
		return nil, ErrTeacherJoin
	}

	var exists bool

	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM class_members
			WHERE class_id = $1 AND user_id = $2
		)
	`, class.ID, userID).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("проверить участие: %w", err)
	}

	if exists {
		return nil, ErrAlreadyMember
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO class_members (class_id, user_id)
		VALUES ($1, $2)
	`, class.ID, userID)

	if err != nil {
		return nil, fmt.Errorf("добавить ученика: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("сохранить участие: %w", err)
	}

	return &class, nil
}
