
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FrontendAPIHandler struct {
	db *pgxpool.Pool
}

func NewFrontendAPIHandler(db *pgxpool.Pool) *FrontendAPIHandler {
	return &FrontendAPIHandler{db: db}
}

func sendJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("JSON:", err)
	}
}

func apiID(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	n, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || n <= 0 {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return 0, false
	}
	return n, true
}

func authID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := AuthenticatedUserID(r.Context())
	if !ok || id <= 0 {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return 0, false
	}
	return id, true
}

func apiFail(w http.ResponseWriter, err error) {
	log.Println("API:", err)
	http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
}

func (h *FrontendAPIHandler) teacherClassAccess(ctx context.Context, teacherID, classID int64) bool {
	var allowed bool
	err := h.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM classes c
			JOIN users u ON u.id=c.teacher_id
			WHERE c.id=$1
				AND c.teacher_id=$2
				AND u.role='teacher'
				AND u.teacher_verified=TRUE
		)
	`, classID, teacherID).Scan(&allowed)
	return err == nil && allowed
}

func (h *FrontendAPIHandler) studentClassAccess(ctx context.Context, userID, classID int64) bool {
	var allowed bool
	err := h.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM class_members
			WHERE class_id=$1 AND user_id=$2
		)
	`, classID, userID).Scan(&allowed)
	return err == nil && allowed
}

func (h *FrontendAPIHandler) summary(ctx context.Context, userID int64) (map[string]any, error) {
	var studied, known, tests, avg, duels, won int

	err := h.db.QueryRow(ctx, `
		SELECT COUNT(*)::int,
			COUNT(*) FILTER (WHERE grade='good')::int
		FROM card_reviews
		WHERE user_id=$1
	`, userID).Scan(&studied, &known)
	if err != nil {
		return nil, err
	}

	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*)::int,
			COALESCE(AVG(percent),0)::int
		FROM test_attempts
		WHERE user_id=$1 AND status='finished'
	`, userID).Scan(&tests, &avg)
	if err != nil {
		return nil, err
	}

	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*)::int,
			COUNT(*) FILTER (WHERE d.winner_user_id=$1)::int
		FROM duel_players dp
		JOIN duels d ON d.id=dp.duel_id
		WHERE dp.user_id=$1 AND d.status='finished'
	`, userID).Scan(&duels, &won)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"cards_studied":    studied,
		"cards_known":      known,
		"tests_taken":      tests,
		"avg_test_percent": avg,
		"duels_played":     duels,
		"duels_won":        won,
	}, nil
}

func (h *FrontendAPIHandler) MyStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := authID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := h.summary(ctx, userID)
	if err != nil {
		apiFail(w, err)
		return
	}

	var total int
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM cards
	`).Scan(&total)
	if err != nil {
		apiFail(w, err)
		return
	}

	stats["cards_total"] = total

	rows, err := h.db.Query(ctx, `
		SELECT d.id, d.title,
			COUNT(c.id)::int,
			COUNT(cr.card_id)::int,
			COUNT(cr.card_id) FILTER (WHERE cr.grade='good')::int
		FROM decks d
		LEFT JOIN cards c ON c.deck_id=d.id
		LEFT JOIN card_reviews cr
			ON cr.card_id=c.id AND cr.user_id=$1
		GROUP BY d.id, d.title
		ORDER BY d.id
	`, userID)
	if err != nil {
		apiFail(w, err)
		return
	}

	decks := make([]map[string]any, 0)

	for rows.Next() {
		var id int64
		var title string
		var total, studied, known int

		if err = rows.Scan(&id, &title, &total, &studied, &known); err != nil {
			break
		}

		decks = append(decks, map[string]any{
			"id":      id,
			"title":   title,
			"total":   total,
			"studied": studied,
			"known":   known,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	stats["decks"] = decks

	rows, err = h.db.Query(ctx, `
		SELECT ta.id, d.id, d.title, ta.percent, ta.finished_at
		FROM test_attempts ta
		JOIN decks d ON d.id=ta.deck_id
		WHERE ta.user_id=$1 AND ta.status='finished'
		ORDER BY ta.finished_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		apiFail(w, err)
		return
	}

	recent := make([]map[string]any, 0)

	for rows.Next() {
		var id, deckID int64
		var title string
		var percent int
		var finished time.Time

		if err = rows.Scan(&id, &deckID, &title, &percent, &finished); err != nil {
			break
		}

		recent = append(recent, map[string]any{
			"id": id,
			"deck": map[string]any{
				"id":    deckID,
				"title": title,
			},
			"percent":     percent,
			"finished_at": finished,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	stats["recent_tests"] = recent
	sendJSON(w, stats)
}

func (h *FrontendAPIHandler) MyClasses(w http.ResponseWriter, r *http.Request) {
	userID, ok := authID(w, r)
	if !ok {
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT c.id, c.name, c.code,
			TRIM(CONCAT(u.first_name, ' ', u.last_name)),
			(
				SELECT COUNT(*)::int
				FROM class_members
				WHERE class_id=c.id
			)
		FROM class_members cm
		JOIN classes c ON c.id=cm.class_id
		JOIN users u ON u.id=c.teacher_id
		WHERE cm.user_id=$1
		ORDER BY c.id
	`, userID)
	if err != nil {
		apiFail(w, err)
		return
	}

	classes := make([]map[string]any, 0)

	for rows.Next() {
		var classID int64
		var name, code, teacher string
		var count int

		if err = rows.Scan(&classID, &name, &code, &teacher, &count); err != nil {
			break
		}

		classes = append(classes, map[string]any{
			"id":           classID,
			"name":         name,
			"code":         code,
			"teacher_name": teacher,
			"member_count": count,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{"classes": classes})
}

func (h *FrontendAPIHandler) ClassInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := authID(w, r)
	if !ok {
		return
	}

	classID, ok := apiID(w, r, "id")
	if !ok {
		return
	}

	if !h.studentClassAccess(r.Context(), userID, classID) {
		http.Error(w, "Класс не найден", http.StatusNotFound)
		return
	}

	var name, teacher string

	err := h.db.QueryRow(r.Context(), `
		SELECT c.name,
			TRIM(CONCAT(u.first_name, ' ', u.last_name))
		FROM classes c
		JOIN users u ON u.id=c.teacher_id
		WHERE c.id=$1
	`, classID).Scan(&name, &teacher)
	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT u.id,
			TRIM(CONCAT(u.first_name, ' ', u.last_name))
		FROM class_members cm
		JOIN users u ON u.id=cm.user_id
		WHERE cm.class_id=$1
		ORDER BY u.id
	`, classID)
	if err != nil {
		apiFail(w, err)
		return
	}

	members := make([]map[string]any, 0)

	for rows.Next() {
		var id int64
		var memberName string

		if err = rows.Scan(&id, &memberName); err != nil {
			break
		}

		members = append(members, map[string]any{
			"id":    id,
			"name":  memberName,
			"is_me": id == userID,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	stats, err := h.summary(r.Context(), userID)
	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{
		"class": map[string]any{
			"id":           classID,
			"name":         name,
			"teacher_name": teacher,
		},
		"members":     members,
		"my_progress": stats,
	})
}

func (h *FrontendAPIHandler) TeacherClass(w http.ResponseWriter, r *http.Request) {
	teacherID, ok := authID(w, r)
	if !ok {
		return
	}

	classID, ok := apiID(w, r, "id")
	if !ok {
		return
	}

	if !h.teacherClassAccess(r.Context(), teacherID, classID) {
		http.Error(w, "Класс не найден", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var name, code string

	err := h.db.QueryRow(ctx, `
		SELECT name, code
		FROM classes
		WHERE id=$1
	`, classID).Scan(&name, &code)
	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT u.id,
			TRIM(CONCAT(u.first_name, ' ', u.last_name)),
			GREATEST(
				u.last_active_at,
				(
					SELECT MAX(reviewed_at)
					FROM card_reviews
					WHERE user_id=u.id
				),
				(
					SELECT MAX(finished_at)
					FROM test_attempts
					WHERE user_id=u.id AND status='finished'
				),
				(
					SELECT MAX(answered_at)
					FROM duel_answers
					WHERE user_id=u.id
				)
			),
			COUNT(DISTINCT cr.card_id)::int,
			COUNT(DISTINCT cr.card_id)
				FILTER (WHERE cr.grade='good')::int
		FROM class_members cm
		JOIN users u ON u.id=cm.user_id
		LEFT JOIN card_reviews cr ON cr.user_id=u.id
		WHERE cm.class_id=$1
		GROUP BY u.id
		ORDER BY u.id
	`, classID)
	if err != nil {
		apiFail(w, err)
		return
	}

	students := make([]map[string]any, 0)

	for rows.Next() {
		var id int64
		var studentName string
		var last *time.Time
		var studied, known int

		if err = rows.Scan(
			&id, &studentName, &last, &studied, &known,
		); err != nil {
			break
		}

		students = append(students, map[string]any{
			"id":               id,
			"name":             studentName,
			"last_active_at":   last,
			"cards_studied":    studied,
			"cards_known":      known,
			"avg_test_percent": 0,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	for _, student := range students {
		var avg int

		err = h.db.QueryRow(ctx, `
			SELECT COALESCE(AVG(percent),0)::int
			FROM test_attempts
			WHERE user_id=$1 AND status='finished'
		`, student["id"]).Scan(&avg)

		if err != nil {
			apiFail(w, err)
			return
		}

		student["avg_test_percent"] = avg
	}

	sendJSON(w, map[string]any{
		"class": map[string]any{
			"id":           classID,
			"name":         name,
			"code":         code,
			"member_count": len(students),
		},
		"students": students,
	})
}

func (h *FrontendAPIHandler) ClassStats(w http.ResponseWriter, r *http.Request) {
	teacherID, ok := authID(w, r)
	if !ok {
		return
	}

	classID, ok := apiID(w, r, "id")
	if !ok {
		return
	}

	if !h.teacherClassAccess(r.Context(), teacherID, classID) {
		http.Error(w, "Класс не найден", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var count, active, studied, avg int

	err := h.db.QueryRow(ctx, `
		SELECT COUNT(*)::int,
			COUNT(*) FILTER (
				WHERE GREATEST(
					u.last_active_at,
					(
						SELECT MAX(reviewed_at)
						FROM card_reviews
						WHERE user_id=u.id
					),
					(
						SELECT MAX(finished_at)
						FROM test_attempts
						WHERE user_id=u.id
							AND status='finished'
					),
					(
						SELECT MAX(answered_at)
						FROM duel_answers
						WHERE user_id=u.id
					)
				) >= NOW() - INTERVAL '7 days'
			)::int
		FROM class_members cm
		JOIN users u ON u.id=cm.user_id
		WHERE cm.class_id=$1
	`, classID).Scan(&count, &active)
	if err != nil {
		apiFail(w, err)
		return
	}

	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM card_reviews cr
		JOIN class_members cm ON cm.user_id=cr.user_id
		WHERE cm.class_id=$1
	`, classID).Scan(&studied)
	if err != nil {
		apiFail(w, err)
		return
	}

	err = h.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(ta.percent),0)::int
		FROM test_attempts ta
		JOIN class_members cm ON cm.user_id=ta.user_id
		WHERE cm.class_id=$1 AND ta.status='finished'
	`, classID).Scan(&avg)
	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT d.id, d.title,
			COUNT(DISTINCT cm.user_id)
				FILTER (WHERE cr.card_id IS NOT NULL)::int,
			COALESCE(
				ROUND(
					100.0 * COUNT(cr.card_id)
						FILTER (WHERE cr.grade='good')
					/ NULLIF(COUNT(c.id),0)
				),
				0
			)::int
		FROM decks d
		CROSS JOIN class_members cm
		LEFT JOIN cards c ON c.deck_id=d.id
		LEFT JOIN card_reviews cr
			ON cr.card_id=c.id AND cr.user_id=cm.user_id
		WHERE cm.class_id=$1
		GROUP BY d.id, d.title
		ORDER BY d.id
	`, classID)
	if err != nil {
		apiFail(w, err)
		return
	}

	decks := make([]map[string]any, 0)

	for rows.Next() {
		var id int64
		var title string
		var started, percent int

		if err = rows.Scan(
			&id, &title, &started, &percent,
		); err != nil {
			break
		}

		if started > 0 {
			decks = append(decks, map[string]any{
				"id":                id,
				"title":             title,
				"students_started":  started,
				"avg_known_percent": percent,
			})
		}
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err = h.db.Query(ctx, `
		SELECT c.question, d.title,
			ROUND(
				100.0 * COUNT(*)
					FILTER (WHERE cr.grade='again')
				/ COUNT(*)
			)::int
		FROM card_reviews cr
		JOIN class_members cm ON cm.user_id=cr.user_id
		JOIN cards c ON c.id=cr.card_id
		JOIN decks d ON d.id=c.deck_id
		WHERE cm.class_id=$1
		GROUP BY c.id, c.question, d.title
		HAVING COUNT(*)
			FILTER (WHERE cr.grade='again') > 0
		ORDER BY 3 DESC, c.id
		LIMIT 10
	`, classID)
	if err != nil {
		apiFail(w, err)
		return
	}

	hardest := make([]map[string]any, 0)

	for rows.Next() {
		var question, deckTitle string
		var rate int

		if err = rows.Scan(
			&question, &deckTitle, &rate,
		); err != nil {
			break
		}

		hardest = append(hardest, map[string]any{
			"question":   question,
			"deck_title": deckTitle,
			"wrong_rate": rate,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{
		"students_count":      count,
		"active_students_7d":  active,
		"cards_studied_total": studied,
		"avg_test_percent":    avg,
		"decks":               decks,
		"hardest_cards":       hardest,
	})
}

func (h *FrontendAPIHandler) StudentProgress(w http.ResponseWriter, r *http.Request) {
	teacherID, ok := authID(w, r)
	if !ok {
		return
	}

	classID, ok := apiID(w, r, "id")
	if !ok {
		return
	}

	studentID, ok := apiID(w, r, "studentID")
	if !ok {
		return
	}

	if !h.teacherClassAccess(r.Context(), teacherID, classID) ||
		!h.studentClassAccess(r.Context(), studentID, classID) {
		http.Error(w, "Ученик не найден", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var name string
	var last *time.Time

	err := h.db.QueryRow(ctx, `
		SELECT TRIM(CONCAT(u.first_name, ' ', u.last_name)),
			GREATEST(
				u.last_active_at,
				(
					SELECT MAX(reviewed_at)
					FROM card_reviews
					WHERE user_id=u.id
				),
				(
					SELECT MAX(finished_at)
					FROM test_attempts
					WHERE user_id=u.id
						AND status='finished'
				),
				(
					SELECT MAX(answered_at)
					FROM duel_answers
					WHERE user_id=u.id
				)
			)
		FROM users u
		WHERE u.id=$1
	`, studentID).Scan(&name, &last)
	if err != nil {
		apiFail(w, err)
		return
	}

	stats, err := h.summary(ctx, studentID)
	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT d.id, d.title,
			COUNT(c.id)::int,
			COUNT(cr.card_id)
				FILTER (WHERE cr.grade='good')::int
		FROM decks d
		LEFT JOIN cards c ON c.deck_id=d.id
		LEFT JOIN card_reviews cr
			ON cr.card_id=c.id AND cr.user_id=$1
		GROUP BY d.id, d.title
		HAVING COUNT(cr.card_id)>0
		ORDER BY d.id
	`, studentID)
	if err != nil {
		apiFail(w, err)
		return
	}

	decks := make([]map[string]any, 0)

	for rows.Next() {
		var id int64
		var title string
		var total, known int

		if err = rows.Scan(
			&id, &title, &total, &known,
		); err != nil {
			break
		}

		decks = append(decks, map[string]any{
			"id":    id,
			"title": title,
			"total": total,
			"known": known,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err = h.db.Query(ctx, `
		SELECT ta.id, d.id, d.title,
			ta.percent, ta.finished_at
		FROM test_attempts ta
		JOIN decks d ON d.id=ta.deck_id
		WHERE ta.user_id=$1 AND ta.status='finished'
		ORDER BY ta.finished_at DESC
		LIMIT 10
	`, studentID)
	if err != nil {
		apiFail(w, err)
		return
	}

	recent := make([]map[string]any, 0)

	for rows.Next() {
		var id, deckID int64
		var title string
		var percent int
		var finished time.Time

		if err = rows.Scan(
			&id, &deckID, &title, &percent, &finished,
		); err != nil {
			break
		}

		recent = append(recent, map[string]any{
			"id": id,
			"deck": map[string]any{
				"id":    deckID,
				"title": title,
			},
			"percent":     percent,
			"finished_at": finished,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	rows, err = h.db.Query(ctx, `
		SELECT c.question, d.title
		FROM card_reviews cr
		JOIN cards c ON c.id=cr.card_id
		JOIN decks d ON d.id=c.deck_id
		WHERE cr.user_id=$1
			AND cr.grade IN ('again', 'hard')
		ORDER BY cr.reviewed_at DESC
		LIMIT 10
	`, studentID)
	if err != nil {
		apiFail(w, err)
		return
	}

	weak := make([]map[string]any, 0)

	for rows.Next() {
		var question, deckTitle string

		if err = rows.Scan(&question, &deckTitle); err != nil {
			break
		}

		weak = append(weak, map[string]any{
			"question":   question,
			"deck_title": deckTitle,
		})
	}

	if err == nil {
		err = rows.Err()
	}
	rows.Close()

	if err != nil {
		apiFail(w, err)
		return
	}

	sendJSON(w, map[string]any{
		"student": map[string]any{
			"id":             studentID,
			"name":           name,
			"last_active_at": last,
		},
		"stats":        stats,
		"decks":        decks,
		"recent_tests": recent,
		"weak_cards":   weak,
	})
}

func (h *FrontendAPIHandler) RemoveStudent(w http.ResponseWriter, r *http.Request) {
	teacherID, ok := authID(w, r)
	if !ok {
		return
	}

	classID, ok := apiID(w, r, "id")
	if !ok {
		return
	}

	studentID, ok := apiID(w, r, "studentID")
	if !ok {
		return
	}

	if !h.teacherClassAccess(r.Context(), teacherID, classID) {
		http.Error(w, "Класс не найден", http.StatusNotFound)
		return
	}

	tag, err := h.db.Exec(r.Context(), `
		DELETE FROM class_members
		WHERE class_id=$1 AND user_id=$2
	`, classID, studentID)
	if err != nil {
		apiFail(w, err)
		return
	}

	if tag.RowsAffected() == 0 {
		http.Error(w, "Ученик не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var _ = errors.Is
var _ = pgx.ErrNoRows
var _ = strings.TrimSpace
