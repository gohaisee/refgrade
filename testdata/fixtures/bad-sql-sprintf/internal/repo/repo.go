package repo

import (
	"database/sql"
	"fmt"
)

func ByEmail(db *sql.DB, email string) error {
	q := fmt.Sprintf("SELECT id FROM users WHERE email = '%s'", email)
	_, err := db.Query(q)
	return err
}
