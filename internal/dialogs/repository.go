package dialogs

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Dialog struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	LastMessage string `json:"last_message"`
}

func (r *Repository) CreateDialog(user1, user2 string) (int, error) {
	var dialogID int

	err := r.db.QueryRow(`
		INSERT INTO dialogs DEFAULT VALUES
		RETURNING id
	`).Scan(&dialogID)
	if err != nil {
		return 0, err
	}

	_, err = r.db.Exec(`
		INSERT INTO dialog_members(dialog_id, username)
		VALUES ($1, $2), ($1, $3)
	`, dialogID, user1, user2)

	return dialogID, err
}

func (r *Repository) GetDialogs(username string) ([]Dialog, error) {
	rows, err := r.db.Query(`
		SELECT
			d.id,
			COALESCE(
				(
					SELECT dm2.username
					FROM dialog_members dm2
					WHERE dm2.dialog_id = d.id AND dm2.username <> $1
					LIMIT 1
				),
				'Dialog'
			) AS title,
			COALESCE(
				(
					SELECT m.text
					FROM messages m
					WHERE m.dialog_id = d.id
					ORDER BY m.id DESC
					LIMIT 1
				),
				''
			) AS last_message
		FROM dialogs d
		JOIN dialog_members dm ON dm.dialog_id = d.id
		WHERE dm.username = $1
		ORDER BY d.id DESC
	`, username)
	if err != nil {
		return []Dialog{}, err
	}
	defer rows.Close()

	result := make([]Dialog, 0)

	for rows.Next() {
		var d Dialog
		if err := rows.Scan(&d.ID, &d.Title, &d.LastMessage); err != nil {
			continue
		}
		result = append(result, d)
	}

	return result, nil
}

func (r *Repository) IsMember(dialogID int, username string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM dialog_members
			WHERE dialog_id=$1 AND username=$2
		)
	`, dialogID, username).Scan(&exists)

	return exists, err
}

func (r *Repository) GetOrCreateDialog(user1, user2 string) (int, error) {
	var id int

	err := r.db.QueryRow(`
		SELECT dm1.dialog_id
		FROM dialog_members dm1
		JOIN dialog_members dm2 ON dm1.dialog_id = dm2.dialog_id
		WHERE dm1.username = $1 AND dm2.username = $2
		LIMIT 1
	`, user1, user2).Scan(&id)

	if err == sql.ErrNoRows {
		return r.CreateDialog(user1, user2)
	}

	if err != nil {
		return 0, err
	}

	return id, nil
}