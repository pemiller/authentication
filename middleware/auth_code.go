package middleware

import (
	"context"
	"fmt"
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
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.PrepareErrorResponse(fmt.Sprintf("Request header is missing authorization with type %s", helpers.AuthTypeCode), nil))
		return
	}

	authCode, err := datastore.GetFromContext(c).GetAuthCode(code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get AuthCode", err))
		return
	}
	if authCode == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("Cannot find AuthCode", nil))
		return
	}

	app := GetApplication(c)
	if app.ID != authCode.ApplicationID {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("AuthCode did not match application", nil))
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
