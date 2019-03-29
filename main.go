package main

import (
	"log"
	"pemiller/authentication/config"
	"pemiller/authentication/handlers/routes"
	"pemiller/authentication/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Parse()

	r := gin.Default()
	r.Use(middleware.SetupDataStore())
	registerRoutes(r)

	err := r.Run(config.Address)
	if err != nil {
		log.Fatal(err)
	}
}

func registerRoutes(e *gin.Engine) {
	api := e.Group("/api")

	app := api.Group("/", middleware.ProcessApplicationHeader)
	app.POST("/code", routes.CreateAuthCode)
	app.GET("/code", middleware.ProcessAuthCodeHeader, routes.GetAuthCode)
	app.DELETE("/code", middleware.ProcessAuthCodeHeader, routes.DeleteAuthCode)

	app.POST("/token", middleware.ProcessAuthCodeHeader, routes.CreateAccessToken)
	app.GET("/token", middleware.ProcessAccessTokenHeader, routes.GetAccessToken)
	app.DELETE("/token", middleware.ProcessAccessTokenHeader, routes.DeleteAccessToken)
	//TODO: restrict this to internal ips only
	app.POST("/token/application", routes.CreateApplicationAccessToken)
}
