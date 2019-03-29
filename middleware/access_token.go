package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"pemiller/authentication/datastore"
	"pemiller/authentication/helpers"
	"pemiller/authentication/models"
)

const SiteHeaderKey = "X-Site"
const AccessTokenHeaderKey = "X-Access-Token"
const accessTokenContextKey = "access_token"

// ProcessAccessTokenHeader checks if the access token header is set in the request with auth type "Token"
// and if so, gets the AccessToken object for that key from the datastore and inserts it into the context
func ProcessAccessTokenHeader(c *gin.Context) {
	token, err := helpers.ParseAuthorizationHeader(c.Request, helpers.AuthTypeToken)
	if len(token) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.PrepareErrorResponse(fmt.Sprintf("Request header is missing authorization with type %s", helpers.AuthTypeToken), nil))
		return
	}

	accessToken, err := datastore.GetFromContext(c).GetAccessToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get AccessToken", err))
		return
	}
	if accessToken == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("Cannot find AccessToken", nil))
		return
	}

	authCode, err := datastore.GetFromContext(c).GetAuthCode(accessToken.AuthCode)
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
	c.Set(accessTokenContextKey, accessToken)
	c.Header(AccessTokenHeaderKey, accessToken.Token)
	c.Next()
}

// GetAccessToken gets the AccessToken object from the context
func GetAccessToken(c *gin.Context) *models.AccessToken {
	result, _ := c.Value(accessTokenContextKey).(*models.AccessToken)
	return result
}

// SetAccessToken inserts an AccessToken object into the context
func SetAccessToken(c *gin.Context, accessToken *models.AccessToken) {
	c.Set(accessTokenContextKey, accessToken)
}
