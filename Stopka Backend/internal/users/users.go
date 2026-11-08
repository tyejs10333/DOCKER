
package users

import (
	"context"
 "crypto/subtle"
 "os"
 "strings"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	PhotoURL  *string `json:"photo_url"`
	Role      *string `json:"role"`
	TeacherVerified bool `json:"teacher_verified"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (*User, error) {
	var user User

	err := r.db.QueryRow(ctx, `
		SELECT id, first_name, last_name, photo_url,
		       role, teacher_verified
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhotoURL,
		&user.Role,
		&user.TeacherVerified,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("получение пользователя: %w", err)
	}

	return &user, nil
}

var ErrInvalidTeacherCode = errors.New("неверный код учителя")
var ErrTeacherCodeNotConfigured = errors.New("код учителя не настроен")

func (r *Repository) UpdateRole(ctx context.Context, id int64, role string, inviteCodes ...string) error {
 if role != "student" && role != "teacher" { return fmt.Errorf("недопустимая роль: %s", role) }
 if role == "teacher" {
   var verified bool
   if err := r.db.QueryRow(ctx, `SELECT teacher_verified FROM users WHERE id = $1`, id).Scan(&verified); err != nil { return fmt.Errorf("проверить подтверждение учителя: %w", err) }
   if !verified {
     expected := strings.TrimSpace(os.Getenv("TEACHER_INVITE_CODE"))
     if expected == "" { return ErrTeacherCodeNotConfigured }
     supplied := ""
     if len(inviteCodes) > 0 { supplied = strings.TrimSpace(inviteCodes[0]) }
     if subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) != 1 { return ErrInvalidTeacherCode }
   }
   _, err := r.db.Exec(ctx, `UPDATE users SET role = 'teacher', teacher_verified = TRUE WHERE id = $1`, id)
   return err
 }
 result, err := r.db.Exec(ctx, `UPDATE users SET role = 'student' WHERE id = $1`, id)
 if err != nil { return fmt.Errorf("обновление роли: %w", err) }
 if result.RowsAffected() == 0 { return fmt.Errorf("пользователь не найден") }
 return nil
}

func (r *Repository) GetOrCreateByMAX(
	ctx context.Context,
	maxID int64,
	firstName string,
	lastName string,
	username string,
) (*User, error) {
	if maxID <= 0 {
		return nil, fmt.Errorf("недопустимый MAX ID")
	}

	var user User

	err := r.db.QueryRow(ctx, `
		INSERT INTO users (
			max_user_id,
			first_name,
			last_name,
			username,
			last_seen_at
		)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (max_user_id)
		DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username,
			last_seen_at = NOW()
		RETURNING
			id,
			first_name,
			last_name,
			photo_url,
			role,
			teacher_verified
	`, maxID, firstName, lastName, username).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhotoURL,
		&user.Role,
		&user.TeacherVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("создание или обновление пользователя MAX: %w", err)
	}

	return &user, nil
}
