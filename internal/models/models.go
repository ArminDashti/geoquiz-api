package models

import "time"

// User is a persisted account.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Username     *string   `json:"username"`
	AvatarPath   *string   `json:"avatar_path,omitempty"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Score is a saved flag or map quiz attempt.
type Score struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	QuizType  string    `json:"quiz_type"`
	Correct   int       `json:"correct"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}
