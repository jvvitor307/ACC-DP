package user

import "time"

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type Cache struct {
	Users        []User `json:"users"`
	ActiveUserID string `json:"active_user_id"`
}
