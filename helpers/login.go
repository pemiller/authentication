package helpers

import (
	"pemiller/authentication/models"
	"time"
)

// GetLoginStatus returns LoginStatus based on the parameters
func GetLoginStatus(isValidated bool, dateExpires *time.Time) models.LoginStatus {
	if !isValidated {
		return models.LoginStatusNotValidated
	}
	if dateExpires != nil && dateExpires.Sub(time.Now().UTC()) <= 0 {
		return models.LoginStatusExpired
	}
	return models.LoginStatusOK
}
