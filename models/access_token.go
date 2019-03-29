package models

import "time"

// AccessToken ...
type AccessToken struct {
	Token         string          `json:"token"`
	Type          AccessTokenType `json:"type"`
	ApplicationID string          `json:"application_id"`
	AuthCode      string          `json:"auth_code"`
	SiteID        string          `json:"site_id"`
	DateCreated   time.Time       `json:"date_created"`
}

// AccessTokenType is a specific string type
type AccessTokenType string

// Possible types of access tokens represented as strings
const (
	AccessTokenTypeUser        AccessTokenType = "User"
	AccessTokenTypeApplication AccessTokenType = "Application"
)
