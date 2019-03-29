package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"pemiller/authentication/datastore"
	"pemiller/authentication/models"
)

// TestPassword compares provided password to password in the User document
func TestPassword(c *gin.Context, user *models.User, providedPassword string) bool {
	var match bool
	if len(user.Password) > 0 {
		match, _ = comparePasswordToHash(user.Password, providedPassword)
	}

	if !match {
		locked, err := datastore.GetFromContext(c).IncrLoginFailCount(user.Email)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return false
		}
		if locked {
			c.String(http.StatusUnauthorized, "Locked")
			return false
		}

		c.Status(http.StatusUnauthorized)
		return false
	}

	return true
}

func comparePasswordToHash(storedPass, providedPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(providedPass))
	return err == nil, nil
}

// CryptPassword creates a hash of the password
func CryptPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
