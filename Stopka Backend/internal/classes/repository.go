package classes

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Class struct {
	ID        int64     `json:"id"`
	TeacherID int64     `json:"teacher_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

var (
	ErrNotTeacher      = errors.New("создавать классы может только учитель")
	ErrNotClassTeacher = errors.New("класс не найден или принадлежит другому учителю")
)

func generateCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const length = 6

	var code strings.Builder

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		code.WriteByte(alphabet[n.Int64()])
	}

	return code.String(), nil
}

func (r *Repository) CreateClass(
	ctx context.Context,
	teacherID int64,
	name string,
) (*Class, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, errors.New("название класса не может быть пустым")
	}

	var isTeacher bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
			  AND role = 'teacher' AND teacher_verified = TRUE
		)
	`, teacherID).Scan(&isTeacher)

	if err != nil {
		return nil, fmt.Errorf("проверить роль пользователя: %w", err)
	}

	if !isTeacher {
		return nil, ErrNotTeacher
	}

	for i := 0; i < 5; i++ {
		code, err := generateCode()
		if err != nil {
			return nil, fmt.Errorf("создать код класса: %w", err)
		}

		var class Class

		err = r.db.QueryRow(ctx, `
			INSERT INTO classes (teacher_id, name, code)
			SELECT id, $2, $3
			FROM users
			WHERE id = $1
			  AND role = 'teacher' AND teacher_verified = TRUE
			ON CONFLICT (code) DO NOTHING
			RETURNING id, teacher_id, name, code, created_at
		`, teacherID, name, code).Scan(
			&class.ID,
			&class.TeacherID,
			&class.Name,
			&class.Code,
			&class.CreatedAt,
		)

		if errors.Is(err, pgx.ErrNoRows) {
			var stillTeacher bool

			checkErr := r.db.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1
					FROM users
					WHERE id = $1
					  AND role = 'teacher' AND teacher_verified = TRUE
				)
			`, teacherID).Scan(&stillTeacher)

			if checkErr != nil {
				return nil, fmt.Errorf("повторно проверить роль: %w", checkErr)
			}

			if !stillTeacher {
				return nil, ErrNotTeacher
			}

			continue
		}

		if err != nil {
			return nil, fmt.Errorf("создать класс: %w", err)
		}

		return &class, nil
	}

	return nil, errors.New("не удалось создать уникальный код класса")
}

type ClassMember struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	JoinedAt  time.Time `json:"joined_at"`
}

func (r *Repository) GetClassMembers(
	ctx context.Context,
	classID int64,
	teacherID int64,
) ([]ClassMember, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM classes
			WHERE id = $1
			  AND teacher_id = $2
		)
	`, classID, teacherID).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("проверить учителя класса: %w", err)
	}

	if !exists {
		return nil, ErrNotClassTeacher
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			cm.joined_at
		FROM class_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.class_id = $1
		ORDER BY cm.joined_at, u.id
	`, classID)

	if err != nil {
		return nil, fmt.Errorf("получить учеников класса: %w", err)
	}
	defer rows.Close()

	members := make([]ClassMember, 0)

	for rows.Next() {
		var member ClassMember

		if err := rows.Scan(
			&member.ID,
			&member.FirstName,
			&member.LastName,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("прочитать ученика: %w", err)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список учеников: %w", err)
	}

	return members, nil
}

type ClassSummary struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func (r *Repository) GetTeacherClasses(
	ctx context.Context,
	teacherID int64,
) ([]ClassSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			c.id,
			c.name,
			c.code,
			COUNT(cm.user_id)::integer AS member_count,
			c.created_at
		FROM classes c
		LEFT JOIN class_members cm ON cm.class_id = c.id
		WHERE c.teacher_id = $1
		GROUP BY c.id, c.name, c.code, c.created_at
		ORDER BY c.created_at DESC, c.id DESC
	`, teacherID)

	if err != nil {
		return nil, fmt.Errorf("получить классы учителя: %w", err)
	}
	defer rows.Close()

	result := make([]ClassSummary, 0)

	for rows.Next() {
		var class ClassSummary

		if err := rows.Scan(
			&class.ID,
			&class.Name,
			&class.Code,
			&class.MemberCount,
			&class.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("прочитать класс: %w", err)
		}

		result = append(result, class)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список классов: %w", err)
	}

	return result, nil
}
