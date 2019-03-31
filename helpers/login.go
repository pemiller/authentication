package helpers

import (
	"time"

	"github.com/pemiller/authentication/models"
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
