package api

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/myhomeworkspace/api-server/errorlog"
	"github.com/whiskeybrav/studentclubportal-server/api/authentication"
	"net/http"
	"strconv"
)

type Post struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Text     string `json:"text"`
	SchoolID int    `json:"school_id"`
	Author   string `json:"author"`
}

type PostsResponse struct {
	Status string `json:"status"`
	Posts  []Post `json:"posts"`
}

func ConfigurePosts(e *echo.Echo) {
	e.GET("/:schoolId/getPosts", func(c echo.Context) error {
		schoolId, err := strconv.Atoi(c.Param("schoolId"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "invalid_params"})
		}

		rows, err := db.Query("SELECT p.id, title, date, text, p.schoolId, u.fname, u.lname, u.showsLastname FROM posts p INNER JOIN users u on p.authorId = u.id WHERE p.schoolId = ? ", schoolId)
		if err != nil {
			errorlog.LogError("getting posts", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var posts []Post

		for rows.Next() {
			post := Post{}
			var firstname, lastname string
			var showsLastname int
			err := rows.Scan(&post.ID, &post.Title, &post.Date, &post.Text, &post.SchoolID, &firstname, &lastname, &showsLastname)
			if err != nil {
				errorlog.LogError("getting posts", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			if showsLastname == 1 {
				post.Author = firstname + " " + lastname
			} else {
				post.Author = firstname
			}

			posts = append(posts, post)
		}

		return c.JSON(http.StatusOK, PostsResponse{
			Status: "ok",
			Posts:  posts,
		})
	})

	e.POST("/posts/new", func(c echo.Context) error {
		if c.FormValue("title") == "" || c.FormValue("text") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		session := authentication.GetSession(c)

		var schoolId int

		err := db.QueryRow("SELECT id from schools WHERE clubheadId = ? OR facultyadviserId = ?", session.UserID, session.UserID).Scan(&schoolId)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "unauthorized"})
		}

		_, err = db.Exec("INSERT INTO posts (title, schoolId, date, authorId, `text`) VALUES (?, ?, NOW(), ?, ?)", c.FormValue("title"), schoolId, session.UserID, c.FormValue("text"))
		if err != nil {
			errorlog.LogError("adding post", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/posts/delete", func(c echo.Context) error {
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

		var postSchoolId int

		err = db.QueryRow("SELECT schoolId from posts WHERE id = ?", postId).Scan(&postSchoolId)
		if err != nil {
			errorlog.LogError("getting id of post to delete", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		if schoolId != postSchoolId {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "unauthorized"})
		}

		_, err = db.Exec("DELETE FROM posts WHERE id = ?", postId)
		if err != nil {
			errorlog.LogError("deleting post", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
