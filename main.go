package main

import (
	"fmt"
	"strings"

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

	e.HideBanner = true

	fmt.Println(logotype)

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

var logotype = strings.Replace(`
__        ___     _     _              ____
\ \      / / |__ (_)___| | _____ _   _| __ ) _ __ __ ___   _____   :::  ::: ::::==  =======
 \ \ /\ / /| '_ \| / __| |/ / _ \ | | |  _ \| '__/ _f \ \ / / _ \  :::  ::: :::     ===  ===
  \ V  V / | | | | \__ \   <  __/ |_| | |_) | | | (_| |\ V / (_) | :::  :::  :::==  ========
   \_/\_/  |_| |_|_|___/_|\_\___|\__, |____/|_|  \__,_| \_/ \___/  ===  ===     === ===  ===
                                 |___/                             ======== ======  ===  ===

                                 studentclubportal-server
`, "f", "`", -1)
