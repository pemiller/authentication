package models

import "time"

// AuthCode ...
type AuthCode struct {
	Code          string        `json:"code"`
	UserID        string        `json:"user_id"`
	Email         string        `json:"email"`
	ApplicationID string        `json:"application_id"`
	AuthType      AuthTypeValue `json:"auth_type"`
	Sites         []string      `json:"sites"`
	IP            string        `json:"ip"`
	DateCreated   time.Time     `json:"date_created"`
}
