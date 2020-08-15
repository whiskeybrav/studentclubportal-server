package api

import (
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/NoteToScreen/maily-go/maily"
	"github.com/labstack/echo"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"github.com/whiskeybrav/studentclubportal-server/errlog"
	"github.com/whiskeybrav/studentclubportal-server/mail"
	"github.com/whiskeybrav/studentclubportal-server/util"
	"golang.org/x/crypto/bcrypt"
)

type SchoolResponse struct {
	Status string `json:"status"`
	School School `json:"school"`
}

type SchoolsResponse struct {
	Status  string   `json:"status"`
	Schools []School `json:"schools"`
}

type School struct {
	Id              int     `json:"id"`
	DisplayName     string  `json:"display_name"`
	Name            string  `json:"name"`
	Website         string  `json:"website"`
	DonationsRaised float64 `json:"donations_raised"`
	DonationGoal    float64 `json:"donation_goal"`
	FoundedDate     string  `json:"founded_date"`
	City            string  `json:"city"`
	State           string  `json:"state"`
	Address         string  `json:"address"`
	DriveFolder     string  `json:"drive_folder"`
	ClubHead        User    `json:"club_head"`
	FacultyAdviser  User    `json:"faculty_adviser"`
	IsVerified      bool    `json:"is_verified"`
}

type User struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	GradeLevel int    `json:"grade_level"`
}

type UsersResponse struct {
	Status string `json:"status"`
	Users  []User `json:"users"`
}

func ConfigureSchools(e *echo.Echo) {
	e.POST("/schools/register", func(c echo.Context) error {
		if c.FormValue("fname") == "" || c.FormValue("lname") == "" || c.FormValue("email") == "" || c.FormValue("password") == "" {
			// These are the things that we need to register the teacher
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		if c.FormValue("displayname") == "" || !isLetter(c.FormValue("displayname")) || c.FormValue("name") == "" || c.FormValue("website") == "" || c.FormValue("city") == "" || len(c.FormValue("state")) != 2 || c.FormValue("address") == "" {
			// These are the things that we need to register the school
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

		rows, err := db.Query("SELECT * FROM schools WHERE displayname = ?", c.FormValue("displayname"))
		if err != nil {
			errlog.LogError("seeing if display name is used", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		if rows.Next() {
			// display name used :'(
			return c.JSON(http.StatusConflict, ErrorResponse{"error", "display_name_already_used"})
		}
		err = rows.Close()
		if err != nil {
			// I think this error would be non-fatal
			errlog.LogError("closing rows (nonfatal)", err)
			err = nil
		}

		_, err = db.Exec("INSERT INTO schools (displayname, name, clubheadId, facultyadviserId, website, donationsRaised, foundedDate, city, state, address, driveFolder, donationGoal, isVerified) VALUES (?, ?, -1, -1, ?, 0, NOW(), ?, ?, ?, ?, 0, -1)",
			strings.ToLower(c.FormValue("displayname")),
			c.FormValue("name"),
			c.FormValue("website"),
			c.FormValue("city"),
			c.FormValue("state"),
			c.FormValue("address"),
			"https://drive.google.com",
		)
		if err != nil {
			errlog.LogError("creating school", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		schoolId := 0

		err = db.QueryRow("SELECT id FROM schools WHERE displayname = ?", c.FormValue("displayname")).Scan(&schoolId)
		if err != nil {
			errlog.LogError("getting id of new school", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		pwd, err := bcrypt.GenerateFromPassword([]byte(c.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			errlog.LogError("generating password hash", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		_, err = db.Exec("INSERT INTO users (fname, lname, email, password, schoolId, type, userLevel, registration) VALUES (?, ?, ?, ?, ?, ?, 0, NOW())", c.FormValue("fname"), c.FormValue("lname"), c.FormValue("email"), string(pwd), schoolId, UserTypeTeacher)
		if err != nil {
			errlog.LogError("adding user to db", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		session := c.Get("session").(authentication.SessionInfo)

		err = db.QueryRow("SELECT id FROM users WHERE email = ?", c.FormValue("email")).Scan(&session.UserID)
		if err != nil {
			errlog.LogError("getting new user id from DB", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		err = authentication.SetSession(session)
		if err != nil {
			errlog.LogError("setting new user's id", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		_, err = db.Exec("UPDATE schools SET facultyadviserId = ? WHERE displayname = ?", session.UserID, c.FormValue("displayname"))

		_, err = mail.Mail.SendMail(config.Mail.AdminName, config.Mail.AdminEmail, "newSchool", maily.TemplateData{}, maily.FuncMap{}, maily.FuncMap{})
		if err != nil {
			errlog.LogError("sending admin registration email", err)
		}

		return statusOk(c)
	})

	e.POST("/schools/setClubHead", func(c echo.Context) error {
		newClubHeadId, err := strconv.Atoi(c.FormValue("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		session := authentication.GetSession(c)

		var schoolId int

		err = db.QueryRow("SELECT id from schools WHERE facultyadviserId = ?", session.UserID).Scan(&schoolId)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "unauthorized"})
		}

		_, err = db.Exec("UPDATE schools SET clubheadId = ? WHERE id = ?", newClubHeadId, schoolId)
		if err != nil {
			errlog.LogError("updating club head", err)
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "internal_server_error"})
		}

		return statusOk(c)
	})

	e.GET("/:schoolId/getMembers", func(c echo.Context) error {
		schoolId, err := strconv.Atoi(c.Param("schoolId"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "invalid_params"})
		}

		rows, err := db.Query("SELECT id, fname, lname, showsLastname, gradeLevel FROM users WHERE schoolId = ?", schoolId)
		if err != nil {
			errlog.LogError("getting events", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		defer rows.Close()

		var users []User

		for rows.Next() {
			user := User{}
			fname := ""
			lname := ""
			showsLName := 0

			err := rows.Scan(&user.Id, &fname, &lname, &showsLName, &user.GradeLevel)
			if err != nil {
				errlog.LogError("scanning event", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			if showsLName == 1 {
				user.Name = fname + " " + lname
			} else {
				user.Name = fname
			}

			users = append(users, user)
		}

		return c.JSON(http.StatusOK, UsersResponse{"ok", users})
	})

	e.GET("/schools/get/:name", func(c echo.Context) error {
		row := db.QueryRow("SELECT s.id, s.displayname, s.name, s.website, s.donationsRaised, s.donationGoal, s.foundedDate, s.city, s.state, s.address, s.driveFolder, s.isVerified, s.clubheadId, ch.fname, ch.showsLastname, ch.lname, ch.gradeLevel, fa.id, fa.fname, fa.lname FROM schools s LEFT OUTER JOIN users ch ON s.clubheadId = ch.id INNER JOIN users fa ON s.facultyadviserId = fa.id WHERE displayname = ?;", c.Param("name"))

		facultyAdviser := User{GradeLevel: -1}
		clubHead := User{}
		school := School{}

		var clubHeadFName, clubHeadLName, clubHeadLNameShown, clubHeadGradeLevel []byte

		adviserFName := ""
		adviserLName := ""

		isVerifiedInt := 0

		err := row.Scan(
			&school.Id,
			&school.DisplayName,
			&school.Name,
			&school.Website,
			&school.DonationsRaised,
			&school.DonationGoal,
			&school.FoundedDate,
			&school.City,
			&school.State,
			&school.Address,
			&school.DriveFolder,
			&isVerifiedInt,
			&clubHead.Id,
			&clubHeadFName,
			&clubHeadLNameShown,
			&clubHeadLName,
			&clubHeadGradeLevel,
			&facultyAdviser.Id,
			&adviserFName,
			&adviserLName,
		)

		if err != nil {
			return c.JSON(http.StatusNotFound, checkErr(err, "invalid_school"))
		}

		facultyAdviser.Name = adviserFName + " " + adviserLName
		chln, _ := strconv.Atoi(string(clubHeadLNameShown))
		chgl, _ := strconv.Atoi(string(clubHeadGradeLevel))

		if chln == 1 {
			clubHead.Name = string(clubHeadFName) + " " + string(clubHeadLName)
		} else {
			clubHead.Name = string(clubHeadFName)
		}

		clubHead.GradeLevel = chgl

		school.IsVerified = isVerifiedInt == 1

		school.ClubHead = clubHead
		school.FacultyAdviser = facultyAdviser

		userID := c.Get("session").(authentication.SessionInfo).UserID

		if school.IsVerified {
			return c.JSON(http.StatusOK, SchoolResponse{"ok", school})
		}

		if userID == -1 || (userID != school.ClubHead.Id && userID != school.FacultyAdviser.Id) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "school_unverified"})
		}

		return c.JSON(http.StatusOK, SchoolResponse{"ok", school})
	})

	e.GET("/schools/search", func(c echo.Context) error {
		q := c.FormValue("q")
		if len(q) <= 2 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "search_query_too_short"})
		}

		// sanitize the query, see https://githubengineering.com/like-injection/ for details
		q = strings.Replace(q, "\\", "\\\\", -1)
		q = strings.Replace(q, "%", "\\%", -1)
		q = strings.Replace(q, "_", "\\_", -1)
		q = "%" + q + "%"

		rows, err := db.Query("SELECT s.id, s.displayname, s.name, s.website, s.donationsRaised, s.donationGoal, s.foundedDate, s.city, s.state, s.address, s.driveFolder, s.isVerified, u1.id, u1.fname, u1.showsLastname, u1.lname, u1.gradeLevel, u2.id, u2.fname, u2.lname FROM schools s JOIN users u1 ON s.clubheadId = u1.id JOIN users u2 ON s.facultyadviserId = u2.id WHERE s.name LIKE ? OR s.displayname LIKE ?", q, q)
		if err != nil {
			errlog.LogError("searching for schools", err)
			return c.JSON(http.StatusNotFound, checkErr(err, "internal_server_error"))
		}

		defer rows.Close()

		var schools []School
		for rows.Next() {
			school := School{}
			isVerifiedInt := 0
			clubHead := User{}
			clubHeadFName := ""
			clubHeadLName := ""
			clubHeadLNameShown := 0
			facultyAdviser := User{}
			adviserFName := ""
			adviserLName := ""

			err := rows.Scan(
				&school.Id,
				&school.DisplayName,
				&school.Name,
				&school.Website,
				&school.DonationsRaised,
				&school.DonationGoal,
				&school.FoundedDate,
				&school.City,
				&school.State,
				&school.Address,
				&school.DriveFolder,
				&isVerifiedInt,
				&clubHead.Id,
				&clubHeadFName,
				&clubHeadLNameShown,
				&clubHeadLName,
				&clubHead.GradeLevel,
				&facultyAdviser.Id,
				&adviserFName,
				&adviserLName,
			)

			if err != nil {
				errlog.LogError("searching for schools", err)
				return c.JSON(http.StatusNotFound, checkErr(err, "internal_server_error"))
			}

			facultyAdviser.Name = adviserFName + " " + adviserLName
			if clubHeadLNameShown == 1 {
				clubHead.Name = clubHeadFName + " " + clubHeadLName
			} else {
				clubHead.Name = clubHeadFName
			}

			school.IsVerified = isVerifiedInt == 1

			if !school.IsVerified {
				continue
			}

			school.ClubHead = clubHead
			school.FacultyAdviser = facultyAdviser

			schools = append(schools, school)
		}

		return c.JSON(http.StatusOK, SchoolsResponse{"ok", schools})
	})
}

func isLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
