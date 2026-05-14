package auth

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(username, password string) error {
	_, err := r.db.Exec(
		"INSERT INTO users(username, password) VALUES($1, $2)",
		username, password,
	)
	
	return err
}

type User struct {
	Username string
	Password string
}

func (r *Repository) GetByUsername(username string) (*User, error) {
	u := &User{}

	err := r.db.QueryRow(
		"SELECT username, password FROM users WHERE username=$1",
		username,
	).Scan(&u.Username, &u.Password)

	if err != nil {
		return nil, err
	}

	return u, err
}