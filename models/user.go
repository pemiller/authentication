package models

import "time"

// User ...
type User struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	Password    string                 `json:"pass"`
	Code        string                 `json:"code"`
	IsValidated bool                   `json:"is_validated"`
	SiteRefs    []string               `json:"site_refs,omitempty"`
	SiteLogins  map[string]*SiteLogins `json:"site_logins,omitempty"`
	Logins      []*LoginTime           `json:"logins,omitempty"`
	DateExpires *time.Time             `json:"date_expires,omitempty"`
}

// SiteLogins ...
type SiteLogins struct {
	Logins []*LoginTime `json:"logins,omitempty"`
}

// LoginTime ...
type LoginTime struct {
	Time     time.Time     `json:"time"`
	AuthType AuthTypeValue `json:"auth_type,omitempty"`
	IP       string        `json:"ip,omitempty"`
}

// AuthTypeValue is a specific string type
type AuthTypeValue string

// Possible types of authorization represented as strings
const (
	AuthTypeUser        AuthTypeValue = "User"
	AuthTypeApplication AuthTypeValue = "Application"
)
