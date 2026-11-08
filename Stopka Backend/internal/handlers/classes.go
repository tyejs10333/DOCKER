
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"stopka/internal/classes"
)

type ClassesHandler struct {
	repository *classes.Repository
	db         *pgxpool.Pool
}

func NewClassesHandler(repository *classes.Repository, db *pgxpool.Pool) *ClassesHandler {
	return &ClassesHandler{repository: repository, db: db}
}

func (h *ClassesHandler) verifiedTeacherID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok || userID <= 0 {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return 0, false
	}

	var verified bool
	err := h.db.QueryRow(r.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id=$1
				AND role='teacher'
				AND teacher_verified=TRUE
		)
	`, userID).Scan(&verified)

	if err != nil {
		log.Println("Ошибка проверки прав учителя:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return 0, false
	}

	if !verified {
		http.Error(w, "Нет прав учителя", http.StatusForbidden)
		return 0, false
	}

	return userID, true
}

func (h *ClassesHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.verifiedTeacherID(w, r)
	if !ok {
		return
	}

	var request struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 4096))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "Ожидается один JSON-объект", http.StatusBadRequest)
		return
	}

	request.Name = strings.TrimSpace(request.Name)

	if request.Name == "" || len([]rune(request.Name)) > 40 {
		http.Error(w, "Название класса должно содержать от 1 до 40 символов", http.StatusBadRequest)
		return
	}

	class, err := h.repository.CreateClass(
		r.Context(),
		userID,
		request.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, classes.ErrNotTeacher):
			http.Error(w, "Создавать классы могут только учителя", http.StatusForbidden)
		default:
			log.Println("Ошибка создания класса:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]any{
		"class": class,
	}); err != nil {
		log.Println("Ошибка отправки класса:", err)
	}
}

func (h *ClassesHandler) JoinClass(w http.ResponseWriter, r *http.Request) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var request struct {
		Code string `json:"code"`
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 4096))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "Ожидается один JSON-объект", http.StatusBadRequest)
		return
	}

	request.Code = strings.ToUpper(strings.TrimSpace(request.Code))

	if len(request.Code) != 6 {
		http.Error(w, "Код класса должен содержать 6 символов", http.StatusBadRequest)
		return
	}

	class, err := h.repository.JoinClass(
		r.Context(),
		userID,
		request.Code,
	)

	if err != nil {
		switch {
		case errors.Is(err, classes.ErrNotStudent):
			http.Error(w, "Вступать в классы могут только ученики", http.StatusForbidden)
		case errors.Is(err, classes.ErrClassNotFound):
			http.Error(w, "Класс не найден", http.StatusNotFound)
		case errors.Is(err, classes.ErrAlreadyMember):
			http.Error(w, "Ученик уже состоит в классе", http.StatusConflict)
		case errors.Is(err, classes.ErrTeacherJoin):
			http.Error(w, "Нельзя вступить в собственный класс", http.StatusForbidden)
		default:
			log.Println("Ошибка вступления в класс:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"class": class,
	}); err != nil {
		log.Println("Ошибка отправки класса:", err)
	}
}

func (h *ClassesHandler) GetClassMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.verifiedTeacherID(w, r)
	if !ok {
		return
	}

	classID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || classID <= 0 {
		http.Error(w, "Неверный ID класса", http.StatusBadRequest)
		return
	}

	members, err := h.repository.GetClassMembers(
		r.Context(),
		classID,
		userID,
	)

	if err != nil {
		switch {
		case errors.Is(err, classes.ErrNotClassTeacher):
			http.Error(w, "Класс не найден", http.StatusNotFound)
		default:
			log.Println("Ошибка получения учеников:", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"members": members,
	}); err != nil {
		log.Println("Ошибка отправки списка учеников:", err)
	}
}

func (h *ClassesHandler) GetTeacherClasses(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.verifiedTeacherID(w, r)
	if !ok {
		return
	}

	classesList, err := h.repository.GetTeacherClasses(
		r.Context(),
		userID,
	)

	if err != nil {
		log.Println("Ошибка получения классов:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"classes": classesList,
	}); err != nil {
		log.Println("Ошибка отправки классов:", err)
	}
}
