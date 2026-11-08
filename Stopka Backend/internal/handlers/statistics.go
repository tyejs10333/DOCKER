package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"stopka/internal/tests"
)

type StatisticsHandler struct {
	repository *tests.Repository
}

func NewStatisticsHandler(
	repository *tests.Repository,
) *StatisticsHandler {
	return &StatisticsHandler{repository: repository}
}

func statisticsUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := AuthenticatedUserID(r.Context())
	if !ok || userID <= 0 {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return 0, false
	}

	return userID, true
}

func statisticsError(w http.ResponseWriter, err error) {
	if errors.Is(err, tests.ErrStatisticsForbidden) {
		http.Error(w, "Нет доступа к статистике", http.StatusForbidden)
		return
	}

	log.Println("Ошибка статистики:", err)
	http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
}

func (h *StatisticsHandler) GetDeckResults(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := statisticsUserID(w, r)
	if !ok {
		return
	}

	deckID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	results, err := h.repository.GetDeckResults(
		r.Context(),
		deckID,
		userID,
	)
	if err != nil {
		statisticsError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"deck_id": deckID,
		"results": results,
	}); err != nil {
		log.Println("Ошибка отправки статистики:", err)
	}
}

func (h *StatisticsHandler) GetQuestionStatistics(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := statisticsUserID(w, r)
	if !ok {
		return
	}

	deckID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || deckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	statistics, err := h.repository.GetQuestionStatistics(
		r.Context(),
		deckID,
		userID,
	)
	if err != nil {
		statisticsError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"deck_id":   deckID,
		"questions": statistics,
	}); err != nil {
		log.Println("Ошибка отправки статистики:", err)
	}
}

func (h *StatisticsHandler) GetClassDeckResults(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := statisticsUserID(w, r)
	if !ok {
		return
	}

	classID, err := strconv.ParseInt(r.PathValue("classID"), 10, 64)
	if err != nil || classID <= 0 {
		http.Error(w, "Неверный ID класса", http.StatusBadRequest)
		return
	}

	deckID, err := strconv.ParseInt(r.PathValue("deckID"), 10, 64)
	if err != nil || deckID <= 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	results, err := h.repository.GetClassDeckResults(
		r.Context(),
		classID,
		deckID,
		userID,
	)
	if err != nil {
		statisticsError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(map[string]any{
		"class_id": classID,
		"deck_id":  deckID,
		"results":  results,
	}); err != nil {
		log.Println("Ошибка отправки статистики класса:", err)
	}
}
