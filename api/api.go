package api

import (
	"database/sql"
	"github.com/labstack/echo"
	"github.com/whiskeybrav/studentclubportal-server/configuration"
	"github.com/whiskeybrav/studentclubportal-server/version"
	"net/http"
	"time"
)

type VersionResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

var config *configuration.Config
var db *sql.DB

func statusOk(c echo.Context) error {
	return c.JSON(http.StatusOK, StatusResponse{"ok"})
}

func checkErr(err error, msg string) ErrorResponse {
	if err != nil {
		return ErrorResponse{
			Status: "error",
			Error:  msg,
		}
	} else {
		return ErrorResponse{}
	}
}

func Configure(e *echo.Echo, configuration *configuration.Config, database *sql.DB) {
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, VersionResponse{"ok", version.Version})
	})

	config = configuration
	db = database

	ConfigureAuth(e)
	ConfigureSchools(e)
	ConfigurePosts(e)
	ConfigureEvents(e)

	e.GET("/teapot", func(c echo.Context) error {
		return c.JSON(http.StatusTeapot, ErrorResponse{"error", "requested_body_is_short_and_stout"})
	})
}

func fixTime(timeObj time.Time) string {
	return timeObj.Format("2006-01-02 15:04:05")
}

func fixBool(b bool) int {
	if b {
		return 1
	}
	return 0
}
