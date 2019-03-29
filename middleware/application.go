package middleware

import (
	"fmt"
	"log"
	"net/http"
	"pemiller/authentication/datastore"
	"pemiller/authentication/models"

	"github.com/gin-gonic/gin"
)

const applicationHeaderKey = "X-Application"
const applicationContextKey = "application"

// ProcessApplicationHeader checks if the application header is set in the request and if so,
// gets the application object for that key from the datastore and inserts it into the context
func ProcessApplicationHeader(c *gin.Context) {
	headerValue := c.Request.Header.Get(applicationHeaderKey)
	if len(headerValue) == 0 {
		log.Printf("Request header is missing %s\n", applicationHeaderKey)
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Missing application request header"))
		return
	}

	app, err := datastore.GetFromContext(c).GetApplication(headerValue)
	if err != nil {
		log.Println("Error getting application from datastore")
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if app == nil {
		log.Println("Cannot find application")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Set(applicationContextKey, app)
	c.Next()
}

// GetApplication gets the Application object from the context
func GetApplication(c *gin.Context) *models.Application {
	result, _ := c.Value(applicationContextKey).(*models.Application)
	return result
}

// SetApplication inserts an Application object into the context
func SetApplication(c *gin.Context, app *models.Application) {
	c.Set(applicationContextKey, app)
}
