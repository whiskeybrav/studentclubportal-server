package api

import (
	"github.com/labstack/echo"
	"github.com/myhomeworkspace/api-server/errorlog"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"github.com/whiskeybrav/studentclubportal-server/util"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
)

const (
	UserTypeTeacher = iota
	UserTypeStudent
)

type me struct {
	Id            int    `json:"id"`
	Fname         string `json:"fname"`
	Lname         string `json:"lname"`
	ShowsLastName bool   `json:"shows_last_name"`
	Email         string `json:"email"`
	SchoolId      int    `json:"school_id"`
	Type          int    `json:"type"`
	GradeLevel    int    `json:"grade_level"`
	HowDidYouHear string `json:"how_did_you_hear"`
	UserLevel     int    `json:"user_level"`
	Registration  string `json:"registration"`
}

type MeResponse struct {
	Status string `json:"status"`
	Me     me     `json:"me"`
}

func ConfigureAuth(e *echo.Echo) {
	e.POST("/auth/registerTeacher", func(c echo.Context) error {
		if c.FormValue("fname") == "" || c.FormValue("lname") == "" || c.FormValue("email") == "" || c.FormValue("password") == "" || c.FormValue("schoolId") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		if !util.EmailIsValid(c.FormValue("email")) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_email"})
		}

		if !authentication.ValidatePassword(c.FormValue("password")) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "insecure_password"})
		}

		empty := ""

		acctExistsErr := db.QueryRow("SELECT id FROM users WHERE email = ?", c.FormValue("email")).Scan(&empty)

		if acctExistsErr == nil {
			// the account already exists
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "account_exists"})
		}

		err := db.QueryRow("SELECT displayname FROM schools WHERE id = ?", c.FormValue("schoolId")).Scan(&empty)

		if err != nil {
			// the school doesn't exist
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "school_not_found"})
		}

		pwd, err := bcrypt.GenerateFromPassword([]byte(c.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			errorlog.LogError("generating password hash", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		_, err = db.Exec("INSERT INTO users (fname, lname, email, password, schoolId, type, userLevel, registration) VALUES (?, ?, ?, ?, ?, ?, 0, NOW())", c.FormValue("fname"), c.FormValue("lname"), c.FormValue("email"), string(pwd), c.FormValue("schoolId"), UserTypeTeacher)
		if err != nil {
			errorlog.LogError("adding user to db", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		session := c.Get("session").(authentication.SessionInfo)

		err = db.QueryRow("SELECT id FROM users WHERE email = ?", c.FormValue("email")).Scan(&session.UserID)
		if err != nil {
			errorlog.LogError("getting new user id from DB", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		err = authentication.SetSession(session)
		if err != nil {
			errorlog.LogError("getting new user id from DB", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return statusOk(c)
	})

	e.POST("/auth/registerStudent", func(c echo.Context) error {
		if c.FormValue("fname") == "" || c.FormValue("lname") == "" || c.FormValue("email") == "" || c.FormValue("password") == "" || c.FormValue("schoolId") == "" || c.FormValue("gradeLevel") == "" || c.FormValue("howDidYouHear") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		gradeLevel, err := strconv.Atoi(c.FormValue("gradeLevel"))
		if err != nil || (gradeLevel < 1 || gradeLevel > 12) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		if c.FormValue("showsLastName") != "true" && c.FormValue("showsLastName") != "false" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		if !util.EmailIsValid(c.FormValue("email")) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_email"})
		}

		if !authentication.ValidatePassword(c.FormValue("password")) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "insecure_password"})
		}

		empty := ""

		acctExistsErr := db.QueryRow("SELECT id FROM users WHERE email = ?", c.FormValue("email")).Scan(&empty)

		if acctExistsErr == nil {
			// the account already exists
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "account_exists"})
		}

		err = db.QueryRow("SELECT displayname FROM schools WHERE id = ?", c.FormValue("schoolId")).Scan(&empty)

		if err != nil {
			// the school doesn't exist
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "school_not_found"})
		}

		pwd, err := bcrypt.GenerateFromPassword([]byte(c.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			errorlog.LogError("generating password hash", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		_, err = db.Exec("INSERT INTO users (fname, showsLastname, lname, email, password, schoolId, type, userLevel, gradeLevel, howDidYouHear, registration) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?, NOW())",
			c.FormValue("fname"),
			c.FormValue("showsLastName") == "true",
			c.FormValue("lname"),
			c.FormValue("email"),
			string(pwd),
			c.FormValue("schoolId"),
			UserTypeStudent,
			gradeLevel,
			c.FormValue("howDidYouHear"),
		)
		if err != nil {
			errorlog.LogError("adding user to db", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		session := c.Get("session").(authentication.SessionInfo)

		err = db.QueryRow("SELECT id FROM users WHERE email = ?", c.FormValue("email")).Scan(&session.UserID)
		if err != nil {
			errorlog.LogError("getting new user id from DB", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		err = authentication.SetSession(session)
		if err != nil {
			errorlog.LogError("getting new user id from DB #2", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return statusOk(c)
	})

	e.POST("/auth/login", func(c echo.Context) error {
		if c.FormValue("email") == "" || c.FormValue("password") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		email := c.FormValue("email")
		password := c.FormValue("password")

		passwordHash := ""
		id := -1

		err := db.QueryRow("SELECT password, id FROM users WHERE email = ?", email).Scan(&passwordHash, &id)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "invalid_login"})
		}

		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
		if err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "invalid_login"})
		}

		session := authentication.GetSession(c)
		session.UserID = id
		err = authentication.SetSession(session)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/auth/logout", func(c echo.Context) error {
		if c.Get("session").(authentication.SessionInfo).UserID == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		session := authentication.GetSession(c)

		session.UserID = -1

		err := authentication.SetSession(session)
		if err != nil {
			errorlog.LogError("logging user out", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return statusOk(c)
	})

	e.GET("/auth/me", func(c echo.Context) error {
		session := authentication.GetSession(c)
		uid := session.UserID

		if uid == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		me := me{}
		showsLastNameInt := 0

		err := db.QueryRow("SELECT fname, lname, showsLastname, email, schoolId, type, gradeLevel, howDidYouHear, userLevel, registration FROM users WHERE id = ?", uid).Scan(
			&me.Fname,
			&me.Lname,
			&showsLastNameInt,
			&me.Email,
			&me.SchoolId,
			&me.Type,
			&me.GradeLevel,
			&me.HowDidYouHear,
			&me.UserLevel,
			&me.Registration,
		)

		if err != nil {
			errorlog.LogError("getting user info", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		me.ShowsLastName = showsLastNameInt == 1

		return c.JSON(http.StatusOK, MeResponse{"ok", me})
	})
}
