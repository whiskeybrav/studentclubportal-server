package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/whiskeybrav/studentclubportal-server/api"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"github.com/whiskeybrav/studentclubportal-server/configuration"
	"github.com/whiskeybrav/studentclubportal-server/mail"
)

var config configuration.Config

func main() {
	config = configuration.Configure()
	mail.ConfigureMail(config)
	initializeDatabase()
	defer deinitializeDatabase()
	authentication.Configure(db)

	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(authentication.SessionMiddleware)
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{config.Server.CORS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))

	api.Configure(e, &config, db)

	e.Logger.Fatal(e.Start(config.Server.Address))
}
