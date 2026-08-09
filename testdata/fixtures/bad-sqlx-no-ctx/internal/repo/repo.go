package repo

import "github.com/jmoiron/sqlx"

func Load(db *sqlx.DB, id int) error {
	var name string
	return db.Get(&name, "SELECT name FROM users WHERE id = $1", id)
}
