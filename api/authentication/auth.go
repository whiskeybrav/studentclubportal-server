package authentication

import (
	"database/sql"
	"strings"
)

var db *sql.DB

func Configure(database *sql.DB) {
	db = database
}

func ValidatePassword(password string) bool {
	return strings.ContainsAny(strings.ToLower(password), "abcdefghijklmnopqrstuvwxyz") && strings.ContainsAny(password, "0123456789") && len(password) >= 8
}
