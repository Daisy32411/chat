CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    dialog_id INT NOT NULL REFERENCES dialogs(id) ON DELETE CASCADE,
    username TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_dialog_id ON messages(dialog_id);