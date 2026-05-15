package chat

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) SaveMessage(msg Message) error {
	_, err := r.db.Exec(
		`INSERT INTO messages(
			dialog_id,
			username, 
			text
		) VALUES($1, $2, $1)`,
		msg.DialogID,
		msg.Username,
		msg.Text,
	)

	return err
}

func (r *Repository) GetMessage(dialogID int) ([]Message, error) {
	rows, err := r.db.Query(`
		SELECT 
			id,
			dialog_id,
			username, 
			text
		FROM 
			messages
		WHERE dialog_id=$1
		ORDER BY id ASC
		LIMIT 100
	`, dialogID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var msg Message

		err := rows.Scan(
			&msg.ID,
			&msg.DialogID,
			&msg.Username,
			&msg.Text,
		)
		if err != nil {
			continue
		}

		messages = append(messages, msg)
	}

	return messages, nil
}