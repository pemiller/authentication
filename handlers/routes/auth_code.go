package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/pemiller/authentication/datastore"
	"github.com/pemiller/authentication/helpers"
	"github.com/pemiller/authentication/middleware"
	"github.com/pemiller/authentication/models"

	"github.com/gin-gonic/gin"
)

// CreateAuthCode creates a new AuthCode for the authorization header in the request
func CreateAuthCode(c *gin.Context) {
	// get the authorization header value if it is basic auth
	headerValue, err := helpers.ParseAuthorizationHeader(c.Request, helpers.AuthTypeBasic)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.PrepareErrorResponse("Unable to parse header", err))
		return
	}

	// parse the authorization header into a username and password
	username, password, err := helpers.DecodeBasicCredentials(headerValue)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to decode credentials", err))
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

	// check the credentials provided in the header and get the user object from the data store
	ok, user := checkAuth(c, username, password, form.IP)
	if !ok {
		return
	}

	app := middleware.GetApplication(c)
	authCode := &models.AuthCode{
		Code:          helpers.GenerateAuthCode(),
		UserID:        user.ID,
		Email:         user.Email,
		ApplicationID: app.ID,
		AuthType:      models.AuthTypeUser,
		Sites:         user.SiteRefs,
		IP:            form.IP,
		DateCreated:   time.Now().UTC(),
	}

	// save AuthCode to data store
	err = datastore.GetFromContext(c).UpsertAuthCode(authCode)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to create AuthCode", err))
		return
	}

	model := &models.AuthCodeDetailed{
		Code:        authCode.Code,
		AuthType:    models.AuthTypeUser,
		Status:      helpers.GetLoginStatus(user.IsValidated, user.DateExpires),
		Application: app,
	}
	model.Sites, err = getSites(c, authCode.Sites)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get list of sites", err))
		return
	}

	datastore.GetFromContext(c).UpsertAuthCodeDetailedToCache(model)
	c.Header(middleware.AuthCodeHeaderKey, authCode.Code)
	c.JSON(http.StatusCreated, model)
}

func checkAuth(c *gin.Context, email, pass, ip string) (bool, *models.User) {
	// check if there is a locking document for the email, preventing access
	if locked, err := datastore.GetFromContext(c).UserIsLocked(email); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to check if user is locked", err))
		return false, nil
	} else if locked {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.PrepareErrorResponse("Locked", nil))
		return false, nil
	}

	// get user document from couchbase datastore
	user, err := datastore.GetFromContext(c).GetUserByEmail(email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get user", err))
		return false, nil
	}
	if user == nil {
		c.Status(http.StatusNotFound)
		return false, nil
	}

	// check if the provided password matches the stored one
	match := helpers.TestPassword(c, user, pass)
	if !match {
		return false, nil
	}

	// clear any login failures if they exist
	err = datastore.GetFromContext(c).ClearLoginFailCount(email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to clear failed logins", err))
		return false, nil
	}

	// append a login record to the list of logins
	err = datastore.GetFromContext(c).UpdateLoginDateForAll(user.ID, models.AuthTypeUser, ip)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to update login date", err))
		return false, nil
	}

	return true, user
}

// GetAuthCode returns an AuthCodeDetailed based on the context
func GetAuthCode(c *gin.Context) {
	authCode := middleware.GetAuthCode(c)

	// check to see if the response model is still in the cache
	response, err := datastore.GetFromContext(c).GetAuthCodeDetailedFromCache(authCode.Code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get AuthCode from cache", err))
		return
	}

	// if the response was not in the cache then build it from the AuthCode
	if response == nil {
		// build list of site models
		sites, err := getSites(c, authCode.Sites)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get list of sites", err))
			return
		}

		// get user model
		user, err := datastore.GetFromContext(c).GetUser(authCode.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.PrepareErrorResponse("Unable to get user", err))
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, helpers.PrepareErrorResponse("Cannot find AuthCode", nil))
			return
		}

		app := middleware.GetApplication(c)
		response = &models.AuthCodeDetailed{
			Code:        authCode.Code,
			AuthType:    authCode.AuthType,
			Status:      helpers.GetLoginStatus(user.IsValidated, user.DateExpires),
			Application: app,
			Sites:       sites,
		}
		datastore.GetFromContext(c).UpsertAuthCodeDetailedToCache(response)
	}
	c.Header(middleware.AuthCodeHeaderKey, authCode.Code)
	c.JSON(http.StatusOK, response)
}

// DeleteAuthCode deletes the AuthCode from datastore and cache
func DeleteAuthCode(c *gin.Context) {
	authCode := middleware.GetAuthCode(c)
	datastore.GetFromContext(c).DeleteAuthCode(authCode.Code)
	c.Status(http.StatusNoContent)
}

// getSites returns a list of site models from a list of site ids
func getSites(c context.Context, sites []string) ([]*models.Site, error) {
	siteChan := make(chan *models.Site)
	errChan := make(chan error)
	result := []*models.Site{}

	for _, siteID := range sites {
		go func(s string) {
			site, err := datastore.GetFromContext(c).GetSite(s)
			if err != nil {
				errChan <- err
			}
			if site != nil {
				siteChan <- site
			}
		}(siteID)
	}

	for range sites {
		select {
		case s := <-siteChan:
			if s != nil {
				result = append(result, s)
			}
		case e := <-errChan:
			return nil, e
		}
	}

	return result, nil
}
