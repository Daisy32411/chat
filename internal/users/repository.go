package users

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Search(query, currentUser string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT username
		FROM users
		WHERE username ILIKE $1 AND username <> $2
		ORDER BY username
		LIMIT 10
	`, "%"+query+"%", currentUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0)

	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			continue
		}
		result = append(result, username)
	}

	return result, nil
}