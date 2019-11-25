package api

import (
	"fmt"
	"github.com/btubbs/datetime"
	"github.com/labstack/echo"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"github.com/whiskeybrav/studentclubportal-server/errlog"
	"net/http"
	"strconv"
)

type Event struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Attendance  string `json:"attendance"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
}

type EventsResponse struct {
	Status string  `json:"status"`
	Events []Event `json:"events"`
}

func ConfigureEvents(e *echo.Echo) {
	e.GET("/:schoolId/getEvents", func(c echo.Context) error {
		schoolId, err := strconv.Atoi(c.Param("schoolId"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "invalid_params"})
		}

		rows, err := db.Query("SELECT id, attendance, title, start, end, description FROM events WHERE end > NOW() AND schoolId = ?", schoolId)
		if err != nil {
			errlog.LogError("getting events", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		defer rows.Close()

		var events []Event

		for rows.Next() {
			event := Event{}
			err := rows.Scan(&event.ID, &event.Attendance, &event.Title, &event.Start, &event.End, &event.Description)
			if err != nil {
				errlog.LogError("scanning event", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			events = append(events, event)
		}

		return c.JSON(http.StatusOK, EventsResponse{"ok", events})
	})

	e.GET("/:schoolId/getAllEvents", func(c echo.Context) error {
		schoolId, err := strconv.Atoi(c.Param("schoolId"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "invalid_params"})
		}

		rows, err := db.Query("SELECT id, attendance, title, start, end, description FROM events WHERE schoolId = ?", schoolId)
		if err != nil {
			errlog.LogError("getting events", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var events []Event

		for rows.Next() {
			event := Event{}
			err := rows.Scan(&event.ID, &event.Attendance, &event.Title, &event.Start, &event.End, &event.Description)
			if err != nil {
				errlog.LogError("scanning event", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			events = append(events, event)
		}

		return c.JSON(http.StatusOK, EventsResponse{"ok", events})
	})

	e.POST("/events/new", func(c echo.Context) error {
		if c.FormValue("title") == "" || c.FormValue("description") == "" || c.FormValue("attendance") == "" || c.FormValue("start") == "" || c.FormValue("end") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// the date of a format string according to go's docs should be Mon Jan 2 15:04:05 -0700 MST 2006, and MySQL defines that their dates are in the
		// YYYY-MM-DD hh:mm:ss format, therefore the following format string is used

		startTimeObj, err := datetime.ParseUTC(c.FormValue("start"))
		endTimeObj, err := datetime.ParseUTC(c.FormValue("end"))
		if err != nil {
			// the date is not ISO8601 formatted, so MySQL won't be able to handle it. We therefore need to return invalid_params.
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		session := authentication.GetSession(c)

		var schoolId int

		err = db.QueryRow("SELECT id from schools WHERE clubheadId = ? OR facultyadviserId = ?", session.UserID, session.UserID).Scan(&schoolId)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "unauthorized"})
		}

		startTime := fixTime(startTimeObj)
		endTime := fixTime(endTimeObj)

		_, err = db.Exec("INSERT INTO events (attendance, title, start, end, description, schoolId) VALUES (?, ?, ?, ?, ?, ?)", c.FormValue("attendance"), c.FormValue("title"), startTime, endTime, c.FormValue("description"), schoolId)
		if err != nil {
			errlog.LogError("adding post", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/events/delete", func(c echo.Context) error {
		fmt.Println(c.FormValue("id"))
		postId, err := strconv.Atoi(c.FormValue("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		session := authentication.GetSession(c)

		var schoolId int

		err = db.QueryRow("SELECT id from schools WHERE clubheadId = ? OR facultyadviserId = ?", session.UserID, session.UserID).Scan(&schoolId)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "unauthorized"})
		}

		var eventSchoolId int

		err = db.QueryRow("SELECT schoolId from events WHERE id = ?", postId).Scan(&eventSchoolId)
		if err != nil {
			errlog.LogError("getting id of event to delete", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		if schoolId != eventSchoolId {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "unauthorized"})
		}

		_, err = db.Exec("DELETE FROM events WHERE id = ?", postId)
		if err != nil {
			errlog.LogError("deleting event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
