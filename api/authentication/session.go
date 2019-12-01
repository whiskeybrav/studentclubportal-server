package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type SessionInfo struct {
	UserID int
	Token  string
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func GenerateSessionToken() (string, error) {
	return GenerateRandomString(26)
}

func GetSessionFromToken(token string) (SessionInfo, error) {
	si := SessionInfo{UserID: -1, Token: token}
	err := db.QueryRow("SELECT userId from sessions WHERE token = ?", token).Scan(&si.UserID)
	return si, err
}

func GetSession(c echo.Context) SessionInfo {
	return c.Get("session").(SessionInfo)
}

func SetSession(info SessionInfo) error {
	id := 0
	err := db.QueryRow("SELECT id from sessions WHERE token = ?", info.Token).Scan(&id)

	if err != nil {
		// We need to create a new session
		// TODO: Implement token expiration from the backend
		_, err := db.Exec("INSERT INTO sessions (token, userId, expiry) VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 1 WEEK))", info.Token, info.UserID)
		return err
	}

	_, err = db.Exec("UPDATE sessions SET userId = ? WHERE token = ?", info.UserID, info.Token)
	return err
}

func SessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("token")
		if err != nil {
			// newToken doesn't exist
			token, err := GenerateSessionToken()
			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
			err = SetSession(SessionInfo{-1, token})
			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			newToken := new(http.Cookie)
			newToken.Name = "token"
			newToken.Value = token
			newToken.Expires = time.Now().Add(24 * 7 * time.Hour)
			newToken.HttpOnly = true
			newToken.Path = "/"
			c.SetCookie(newToken)

			c.Set("session", SessionInfo{-1, token})

			return next(c)
		}

		token := cookie.Value

		session, err := GetSessionFromToken(token)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		c.Set("session", session)

		return next(c)
	}
}
