package repo

import (
	"database/sql"
	"fmt"
)

func ByTable(db *sql.DB, tableName string) error {
	q := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", tableName)
	_, err := db.Query(q)
	return err
}
