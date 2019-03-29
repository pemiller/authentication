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

	log.Printf("listening on '%s'\n", config.Address)

	err := r.Run(config.Address)
	if err != nil {
		log.Fatal(err)
	}
}

func registerRoutes(e *gin.Engine) {
	api := e.Group("/api")

	app := api.Group("/", middleware.ProcessApplicationHeader)
	app.POST("/", routes.CreateAuthCode)
	app.GET("/", middleware.ProcessAuthCodeHeader, routes.GetAuthCode)
	app.DELETE("/", middleware.ProcessAuthCodeHeader, routes.DeleteAuthCode)
}
