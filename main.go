package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/whiskeybrav/studentclubportal-server/api"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"github.com/whiskeybrav/studentclubportal-server/configuration"
)

var config configuration.Config

func main() {
	config = configuration.Configure()
	initializeDatabase()
	defer deinitializeDatabase()
	authentication.Configure(db)

	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(authentication.SessionMiddleware)
	e.Use(middleware.RequestID())

	api.Configure(e, &config, db)

	e.Logger.Fatal(e.Start(config.Server.Address))
}
