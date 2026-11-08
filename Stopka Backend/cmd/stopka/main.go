package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"stopka/internal/classes"
	"stopka/internal/config"
	"stopka/internal/database"
	"stopka/internal/decks"
	"stopka/internal/handlers"
	"stopka/internal/tests"
	"stopka/internal/users"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка конфигурации: ", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных: ", err)
	}
	defer db.Close()

	log.Println("Подключение к PostgreSQL успешно!")

	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Ошибка миграций: ", err)
	}
	log.Println("Структура базы данных готова!")

	deckRepository := decks.NewRepository(db)
	testRepository := tests.NewRepository(db)
	classesRepository := classes.NewRepository(db)
	userRepository := users.NewRepository(db)

	decksHandler := handlers.NewDecksHandler(deckRepository)
	cardsHandler := handlers.NewCardsHandler(deckRepository)
	questionsHandler := handlers.NewQuestionsHandler(deckRepository)
	answersHandler := handlers.NewAnswersHandler(deckRepository)

	testHistoryHandler := handlers.NewTestHistoryHandler(testRepository)
	testDetailHandler := handlers.NewTestDetailHandler(testRepository)
	statisticsHandler := handlers.NewStatisticsHandler(testRepository)
	testAttemptsHandler := handlers.NewTestAttemptsHandler(testRepository, deckRepository)
	testAnswersHandler := handlers.NewTestAnswersHandler(testRepository)
	testFinishHandler := handlers.NewTestFinishHandler(testRepository, deckRepository)

	classesHandler := handlers.NewClassesHandler(classesRepository, db)
	usersHandler := handlers.NewUsersHandler(userRepository)
	cardReviewsHandler := handlers.NewCardReviewsHandler(db)
	duelsHandler := handlers.NewDuelsHandler(db)
	frontendHandler := handlers.NewFrontendAPIHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /api/decks", decksHandler.GetAll)
	mux.HandleFunc("GET /api/decks/{id}/cards", cardsHandler.GetCards)
	mux.HandleFunc("GET /api/decks/{id}/questions", questionsHandler.GetQuestions)

	mux.HandleFunc("GET /api/me", usersHandler.GetMe)
	mux.HandleFunc("GET /api/me/stats", frontendHandler.MyStats)
	mux.HandleFunc("GET /api/classes", frontendHandler.MyClasses)
	mux.HandleFunc("GET /api/classes/{id}", frontendHandler.ClassInfo)
	mux.HandleFunc("GET /api/teacher/classes/{id}", frontendHandler.TeacherClass)
	mux.HandleFunc("GET /api/teacher/classes/{id}/stats", frontendHandler.ClassStats)
	mux.HandleFunc("GET /api/teacher/classes/{id}/students/{studentID}", frontendHandler.StudentProgress)
	mux.HandleFunc("DELETE /api/teacher/classes/{id}/students/{studentID}", frontendHandler.RemoveStudent)
	mux.HandleFunc("PUT /api/me/role", usersHandler.UpdateRole)

	mux.HandleFunc("POST /api/tests/check-answer", answersHandler.CheckAnswer)
	mux.HandleFunc("POST /api/tests/attempts", testAttemptsHandler.CreateAttempt)
	mux.HandleFunc("POST /api/tests/attempts/{id}/answers", testAnswersHandler.SaveAnswer)
	mux.HandleFunc("POST /api/tests/attempts/{id}/finish", testFinishHandler.FinishAttempt)

	mux.HandleFunc("POST /api/classes", classesHandler.CreateClass)
	mux.HandleFunc("POST /api/teacher/classes", classesHandler.CreateClass)
	mux.HandleFunc("POST /api/classes/join", classesHandler.JoinClass)
	mux.HandleFunc("GET /api/classes/{id}/members", classesHandler.GetClassMembers)

	mux.HandleFunc("GET /api/teacher/classes", classesHandler.GetTeacherClasses)
	mux.HandleFunc("GET /api/teacher/decks/{id}/results", statisticsHandler.GetDeckResults)
	mux.HandleFunc("GET /api/teacher/decks/{id}/question-statistics", statisticsHandler.GetQuestionStatistics)
	mux.HandleFunc("GET /api/teacher/classes/{classID}/decks/{deckID}/results", statisticsHandler.GetClassDeckResults)

	mux.HandleFunc("POST /api/study/reviews", cardReviewsHandler.SaveReview)

	mux.HandleFunc("POST /api/tests", testAttemptsHandler.CreateAttempt)
	mux.HandleFunc("POST /api/tests/{id}/answers", testAnswersHandler.SaveAnswer)
	mux.HandleFunc("POST /api/tests/{id}/finish", testFinishHandler.FinishAttempt)
	mux.HandleFunc("GET /api/tests", testHistoryHandler.GetHistory)
	mux.HandleFunc("GET /api/tests/{id}", testDetailHandler.GetDetail)

	mux.HandleFunc("GET /api/duels", duelsHandler.ListFrontend)
	mux.HandleFunc("POST /api/duels", duelsHandler.Create)
	mux.HandleFunc("POST /api/duels/join", duelsHandler.Join)
	mux.HandleFunc("GET /api/duels/{id}", duelsHandler.GetFrontendGame)
	mux.HandleFunc("POST /api/duels/{id}/answers", duelsHandler.AnswerFrontend)
	mux.HandleFunc("POST /api/duels/{id}/leave", duelsHandler.LeaveFrontend)

	mux.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      loggingMiddleware(handlers.MAXAuthMiddleware(mux, cfg.AppEnv, userRepository)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Сервер «СТОПКА» запущен на http://localhost" + cfg.HTTPAddr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Ошибка HTTP-сервера: ", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Останавливаем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("Ошибка при остановке сервера: ", err)
	}

	log.Println("Сервер остановлен.")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
