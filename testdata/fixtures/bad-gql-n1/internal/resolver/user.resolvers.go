package resolver

import "database/sql"

type Resolver struct {
	db *sql.DB
}

func (r *Resolver) LoadNames(ids []int) error {
	for _, id := range ids {
		_, err := r.db.Query("SELECT name FROM users WHERE id = $1", id)
		if err != nil {
			return err
		}
	}
	return nil
}
