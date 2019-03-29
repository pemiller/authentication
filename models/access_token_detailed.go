package models

import "time"

// AccessTokenDetailed ...
type AccessTokenDetailed struct {
	Token               string        `json:"token"`
	UserID              string        `json:"user_id"`
	Email               string        `json:"email"`
	IsValidated         bool          `json:"is_validated"`
	DatePasswordExpires *time.Time    `json:"date_password_expires"`
	Site                *Site         `json:"site"`
	Status              LoginStatus   `json:"status"`
	AuthType            AuthTypeValue `json:"auth_type"`
	Application         *Application  `json:"application"`
}
