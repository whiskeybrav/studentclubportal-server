package api

import (
	"github.com/labstack/echo"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"github.com/whiskeybrav/studentclubportal-server/errlog"
	"net/http"
	"strings"
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

func ConfigureSchools(e *echo.Echo) {
	e.GET("/schools/get/:name", func(c echo.Context) error {
		row := db.QueryRow("SELECT s.id, s.displayname, s.name, s.website, s.donationsRaised, s.donationGoal, s.foundedDate, s.city, s.state, s.address, s.driveFolder, s.isVerified, s.clubheadId, ch.fname, ch.showsLastname, ch.lname, ch.gradeLevel, fa.id, fa.fname, fa.lname FROM schools s INNER JOIN users ch on s.clubheadId = ch.id INNER JOIN users fa on s.facultyadviserId = fa.id WHERE displayname = ?;", c.Param("name"))

		facultyAdviser := User{GradeLevel: -1}
		clubHead := User{}
		school := School{}

		clubHeadFName := ""
		clubHeadLName := ""
		clubHeadLNameShown := 0

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
			&clubHead.GradeLevel,
			&facultyAdviser.Id,
			&adviserFName,
			&adviserLName,
		)

		facultyAdviser.Name = adviserFName + " " + adviserLName
		if clubHeadLNameShown == 1 {
			clubHead.Name = clubHeadFName + " " + clubHeadLName
		} else {
			clubHead.Name = clubHeadFName
		}

		school.IsVerified = isVerifiedInt == 1

		school.ClubHead = clubHead
		school.FacultyAdviser = facultyAdviser

		userID := c.Get("session").(authentication.SessionInfo).UserID

		if !school.IsVerified && userID != school.ClubHead.Id && userID != school.FacultyAdviser.Id {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "school_unverified"})
		}

		if (checkErr(err, "invalid_school") != ErrorResponse{}) {
			return c.JSON(http.StatusNotFound, checkErr(err, "invalid_school"))
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

		schools := []School{}
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
