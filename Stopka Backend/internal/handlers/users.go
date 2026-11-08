package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"stopka/internal/users"
)

const (
	maxTeacherRoleAttempts = 5
	teacherRoleWindow      = 15 * time.Minute
)

type teacherRoleAttempt struct {
	count   int
	started time.Time
}

type UsersHandler struct {
	repository   *users.Repository
	attemptsMu   sync.Mutex
	roleAttempts map[int64]teacherRoleAttempt
}

func NewUsersHandler(repository *users.Repository) *UsersHandler {
	return &UsersHandler{
		repository:   repository,
		roleAttempts: make(map[int64]teacherRoleAttempt),
	}
}

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	id, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	user, err := h.repository.GetByID(r.Context(), id)
	if err != nil {
		log.Println("Ошибка получения пользователя:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"user":             user,
		"role":             user.Role,
		"teacher_verified": user.TeacherVerified,
	}); err != nil {
		log.Println("Ошибка отправки пользователя:", err)
	}
}

func (h *UsersHandler) reserveTeacherRoleAttempt(userID int64) bool {
	h.attemptsMu.Lock()
	defer h.attemptsMu.Unlock()

	now := time.Now()
	state, exists := h.roleAttempts[userID]

	if !exists || now.Sub(state.started) >= teacherRoleWindow {
		state = teacherRoleAttempt{started: now}
	}

	if state.count >= maxTeacherRoleAttempts {
		return false
	}

	state.count++
	h.roleAttempts[userID] = state

	return true
}

func (h *UsersHandler) resetTeacherRoleAttempts(userID int64) {
	h.attemptsMu.Lock()
	defer h.attemptsMu.Unlock()

	delete(h.roleAttempts, userID)
}

func (h *UsersHandler) releaseTeacherRoleAttempt(userID int64) {
	h.attemptsMu.Lock()
	defer h.attemptsMu.Unlock()

	state, exists := h.roleAttempts[userID]
	if !exists {
		return
	}

	state.count--

	if state.count <= 0 {
		delete(h.roleAttempts, userID)
		return
	}

	h.roleAttempts[userID] = state
}

func (h *UsersHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, ok := AuthenticatedUserID(r.Context())
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var request struct {
		Role       string `json:"role"`
		InviteCode string `json:"invite_code"`
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 4096))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "Ожидается один JSON-объект", http.StatusBadRequest)
		return
	}

	if request.Role != "student" && request.Role != "teacher" {
		http.Error(w, "Роль должна быть student или teacher", http.StatusBadRequest)
		return
	}

	if request.Role == "teacher" && !h.reserveTeacherRoleAttempt(id) {
		http.Error(
			w,
			"Слишком много попыток. Попробуйте через 15 минут",
			http.StatusTooManyRequests,
		)
		return
	}

	err := h.repository.UpdateRole(
		r.Context(),
		id,
		request.Role,
		request.InviteCode,
	)

	if err != nil {
		if request.Role == "teacher" &&
			!errors.Is(err, users.ErrInvalidTeacherCode) {
			h.releaseTeacherRoleAttempt(id)
		}

		if errors.Is(err, users.ErrInvalidTeacherCode) {
			http.Error(w, "Неверный код учителя", http.StatusForbidden)
			return
		}

		if errors.Is(err, users.ErrTeacherCodeNotConfigured) {
			http.Error(
				w,
				"Код учителя не настроен на сервере",
				http.StatusServiceUnavailable,
			)
			return
		}

		log.Println("Ошибка обновления роли:", err)
		http.Error(w, "Не удалось обновить роль", http.StatusInternalServerError)
		return
	}

	if request.Role == "teacher" {
		h.resetTeacherRoleAttempts(id)
	}

	user, err := h.repository.GetByID(r.Context(), id)
	if err != nil || user == nil {
		log.Println("Ошибка получения пользователя после обновления:", err)
		http.Error(
			w,
			"Ошибка получения пользователя",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"user":             user,
		"role":             user.Role,
		"teacher_verified": user.TeacherVerified,
	}); err != nil {
		log.Println("Ошибка отправки ответа:", err)
	}
}