package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"pemiller/authentication/datastore"
	"pemiller/authentication/helpers"
	"pemiller/authentication/models"
)

const AuthCodeHeaderKey = "X-Auth-Code"
const authCodeContextKey = "auth_code"

// ProcessAuthCodeHeader checks if the authorization header is set in the request with auth type "Code"
// and if so, gets the AuthCode object for that key from the datastore and inserts it into the context
func ProcessAuthCodeHeader(c *gin.Context) {
	code, err := helpers.ParseAuthorizationHeader(c.Request, helpers.AuthTypeCode)
	if len(code) == 0 {
		log.Printf("Request header is missing authorization with type %s\n", helpers.AuthTypeCode)
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Missing authorization request header"))
		return
	}

	authCode, err := datastore.GetFromContext(c).GetAuthCode(code)
	if err != nil {
		log.Println("Error getting auth code from datastore")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if authCode == nil {
		log.Println("Cannot find auth code")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	app := GetApplication(c)
	if app.ID != authCode.ApplicationID {
		log.Println("Auth code did not match application in the context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set(authCodeContextKey, authCode)
	c.Header(AuthCodeHeaderKey, authCode.Code)
	c.Next()
}

// GetAuthCode gets the AuthCode object from the context
func GetAuthCode(c context.Context) *models.AuthCode {
	result, _ := c.Value(authCodeContextKey).(*models.AuthCode)
	return result
}

// SetAuthCode inserts an AuthCode object into the context
func SetAuthCode(c *gin.Context, authCode *models.AuthCode) {
	c.Set(authCodeContextKey, authCode)
}
