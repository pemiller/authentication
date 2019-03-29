package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"pemiller/authentication/datastore"
	"pemiller/authentication/helpers"
	"pemiller/authentication/middleware"
	"pemiller/authentication/models"
)

// CreateAccessToken creates a new AccessToken for the authorization header in the request
func CreateAccessToken(c *gin.Context) {
	authCode := middleware.GetAuthCode(c)

	// get the siteID from the header value
	siteID := c.Request.Header.Get(middleware.SiteHeaderKey)
	if len(siteID) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.PrepareErrorResponse(fmt.Sprintf("Request header is missing %s", middleware.SiteHeaderKey), nil))
		return
	}

	// get user document from couchbase datastore
	user, err := datastore.GetFromContext(c).GetUser(authCode.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get user", err))
		return
	}
	if user == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, helpers.PrepareErrorResponse("User not found", nil))
		return
	}

	// get site from couchbase datastore
	site, err := getSite(c, siteID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get site", err))
		return
	}
	if site == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("Site not found", nil))
		return
	}

	app := middleware.GetApplication(c)
	accessToken := &models.AccessToken{
		Token:         helpers.GenerateAccessToken(authCode.Code),
		Type:          models.AccessTokenTypeUser,
		ApplicationID: app.ID,
		AuthCode:      authCode.Code,
		SiteID:        site.SiteID,
		DateCreated:   time.Now().UTC(),
	}

	// save AccessToken to data store
	err = datastore.GetFromContext(c).UpsertAccessToken(accessToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to create AccessToken", err))
		return
	}

	model := &models.AccessTokenDetailed{
		Token:               accessToken.Token,
		UserID:              user.ID,
		Email:               user.Email,
		IsValidated:         user.IsValidated,
		DatePasswordExpires: user.DateExpires,
		Site:                site,
		Status:              helpers.GetLoginStatus(user.IsValidated, user.DateExpires),
		AuthType:            authCode.AuthType,
		Application:         app,
	}

	datastore.GetFromContext(c).UpsertAccessTokenDetailedToCache(model)
	datastore.GetFromContext(c).UpdateLoginDateForSite(user.ID, site.SiteID, authCode.AuthType, authCode.IP)
	c.Header(middleware.AccessTokenHeaderKey, accessToken.Token)
	c.Header(middleware.SiteHeaderKey, site.SiteURL)
	c.JSON(http.StatusCreated, model)
}

// CreateApplicationAccessToken creates a new access token for an Application and Site in the request header
func CreateApplicationAccessToken(c *gin.Context) {
	// get the siteID from the header value
	siteID := c.Request.Header.Get(middleware.SiteHeaderKey)
	if len(siteID) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.PrepareErrorResponse(fmt.Sprintf("Request header is missing %s", middleware.SiteHeaderKey), nil))
		return
	}

	// get site from couchbase datastore
	site, err := getSite(c, siteID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get site", err))
		return
	}
	if site == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("Site not found", nil))
		return
	}

	// get the ip address from the body if it was provided
	form := &models.CreateAuthCodeRequest{}
	if c.Request.Body != http.NoBody {
		err = c.BindJSON(form)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to read body", err))
			return
		}
	}

	app := middleware.GetApplication(c)
	authCode := &models.AuthCode{
		Code:          helpers.GenerateAuthCode(),
		ApplicationID: app.ID,
		AuthType:      models.AuthTypeApplication,
		Sites:         []string{},
		IP:            form.IP,
		DateCreated:   time.Now().UTC(),
	}

	// save AuthCode to data store
	err = datastore.GetFromContext(c).UpsertAuthCode(authCode)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to create AuthCode", err))
		return
	}

	accessToken := &models.AccessToken{
		Token:         helpers.GenerateAccessToken(authCode.Code),
		Type:          models.AccessTokenTypeApplication,
		ApplicationID: app.ID,
		AuthCode:      authCode.Code,
		SiteID:        site.SiteID,
		DateCreated:   time.Now().UTC(),
	}

	// save AccessToken to data store
	err = datastore.GetFromContext(c).UpsertAccessToken(accessToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to create AccessToken", err))
		return
	}

	model := &models.AccessTokenDetailed{
		Token:       accessToken.Token,
		IsValidated: true,
		Site:        site,
		Status:      helpers.GetLoginStatus(true, nil),
		AuthType:    authCode.AuthType,
		Application: app,
	}

	datastore.GetFromContext(c).UpsertAccessTokenDetailedToCache(model)
	c.Header(middleware.AccessTokenHeaderKey, accessToken.Token)
	c.Header(middleware.SiteHeaderKey, site.SiteURL)
	c.JSON(http.StatusCreated, model)
}

// GetAccessToken returns data related to the access token
func GetAccessToken(c *gin.Context) {
	accessToken := middleware.GetAccessToken(c)

	// check to see if the response model is still in the cache
	response, err := datastore.GetFromContext(c).GetAccessTokenDetailedFromCache(accessToken.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get AccessToken from cache", err))
		return
	}

	// if the response was not in the cache then build it from the AccessToken
	if response == nil {
		app := middleware.GetApplication(c)
		authCode := middleware.GetAuthCode(c)

		if accessToken.Type == models.AccessTokenTypeApplication {
			// NYI
		} else {
			// get user document from couchbase datastore
			user, err := datastore.GetFromContext(c).GetUser(authCode.UserID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get user", err))
				return
			}
			if user == nil {
				c.AbortWithStatusJSON(http.StatusNotFound, helpers.PrepareErrorResponse("User not found", nil))
				return
			}

			// get site from couchbase datastore
			site, err := datastore.GetFromContext(c).GetSite(accessToken.SiteID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get site", err))
				return
			}
			if site == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("Site not found", nil))
				return
			}

			response = &models.AccessTokenDetailed{
				Token:               accessToken.Token,
				UserID:              user.ID,
				Email:               user.Email,
				IsValidated:         user.IsValidated,
				DatePasswordExpires: user.DateExpires,
				Site:                site,
				Status:              helpers.GetLoginStatus(user.IsValidated, user.DateExpires),
				AuthType:            authCode.AuthType,
				Application:         app,
			}
		}

		datastore.GetFromContext(c).UpsertAccessTokenDetailedToCache(response)
	}

	c.Header(middleware.SiteHeaderKey, response.Site.SiteURL)
	c.JSON(http.StatusOK, response)
}

// DeleteAccessToken deletes the access token
func DeleteAccessToken(c *gin.Context) {
	accessToken := middleware.GetAccessToken(c)
	datastore.GetFromContext(c).DeleteAccessToken(accessToken.Token)
	c.Status(http.StatusNoContent)
}

func getSite(c context.Context, siteID string) (*models.Site, error) {
	var err error
	var site *models.Site

	_, err = uuid.Parse(siteID)
	if err == nil {
		site, err = datastore.GetFromContext(c).GetSite(siteID)
	} else {
		site, err = datastore.GetFromContext(c).GetSiteByURL(siteID)
	}

	if err != nil {
		return nil, err
	}
	if site == nil {
		return nil, nil
	}

	return site, nil
}
