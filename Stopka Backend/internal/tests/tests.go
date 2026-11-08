
package tests

import "time"

type Attempt struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	DeckID     int64      `json:"deck_id"`
	Status     string     `json:"status"`
	Correct    int        `json:"correct"`
	Total      int        `json:"total"`
	Percent    int        `json:"percent"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

type Answer struct {
	AttemptID  int64     `json:"attempt_id"`
	QuestionID int64     `json:"question_id"`
	OptionID   int64     `json:"option_id"`
	IsCorrect  bool      `json:"is_correct"`
	AnsweredAt time.Time `json:"answered_at"`
}
