
package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"stopka/internal/maxauth"
	"stopka/internal/users"
)

type authenticatedUserKey struct{}

func AuthenticatedUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(authenticatedUserKey{}).(int64)
	return id, ok && id > 0
}

func MAXAuthMiddleware(
	next http.Handler,
	appEnv string,
	repository *users.Repository,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if appEnv == "dev" {
			rawID := r.Header.Get("X-Dev-User-Id")
			if rawID == "" {
				rawID = "1"
			}

			id, err := strconv.ParseInt(rawID, 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "Неверный тестовый ID", http.StatusBadRequest)
				return
			}

			if !allowStudentDuelAccess(w, r, repository, id) {
				return
			}

			forwardAuthenticatedRequest(next, w, r, id)
			return
		}

		token := os.Getenv("MAX_BOT_TOKEN")
		if token == "" {
			log.Println("MAX_BOT_TOKEN не настроен")
			http.Error(w, "Авторизация недоступна", http.StatusServiceUnavailable)
			return
		}

		initData := r.Header.Get("X-Max-Init-Data")
		maxUser, err := maxauth.Validate(initData, token)
		if err != nil {
			http.Error(w, "Необходима авторизация MAX", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		user, err := repository.GetOrCreateByMAX(
			ctx,
			maxUser.ID,
			maxUser.FirstName,
			maxUser.LastName,
			maxUser.Username,
		)
		if err != nil {
			log.Println("Ошибка авторизации пользователя MAX:", err)
			http.Error(w, "Ошибка авторизации", http.StatusInternalServerError)
			return
		}

		if !allowStudentDuelAccess(w, r, repository, user.ID) {
			return
		}

		forwardAuthenticatedRequest(next, w, r, user.ID)
	})
}

func forwardAuthenticatedRequest(
	next http.Handler,
	w http.ResponseWriter,
	r *http.Request,
	id int64,
) {
	ctx := context.WithValue(r.Context(), authenticatedUserKey{}, id)
	request := r.Clone(ctx)

	request.Header.Set("X-Dev-User-Id", strconv.FormatInt(id, 10))

	next.ServeHTTP(w, request)
}

func allowStudentDuelAccess(
	w http.ResponseWriter,
	r *http.Request,
	repository *users.Repository,
	userID int64,
) bool {
	if r.URL.Path != "/api/duels" &&
		!strings.HasPrefix(r.URL.Path, "/api/duels/") {
		return true
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := repository.GetByID(ctx, userID)
	if err != nil {
		log.Println("Ошибка проверки роли для дуэли:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return false
	}

	if user == nil || user.Role == nil || *user.Role != "student" {
		http.Error(w, "Дуэли доступны только ученикам", http.StatusForbidden)
		return false
	}

	return true
}
