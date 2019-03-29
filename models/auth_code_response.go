package models

// AuthCodeResponse ...
type AuthCodeResponse struct {
	Code        string        `json:"code"`
	AuthType    AuthTypeValue `json:"auth_type"`
	Status      LoginStatus   `json:"status"`
	Sites       []*Site       `json:"sites"`
	Application *Application  `json:"application"`
}

// LoginStatus ...
type LoginStatus string

// Login Information ...
const (
	LoginStatusOK              = "OK"
	LoginStatusExpired         = "PasswordExpired"
	LoginStatusNotValidated    = "NotValidated"
	LoginStatusSiteUnavailable = "SiteUnavailable"
)
