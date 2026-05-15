CREATE TABLE dialogs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE dialog_members (
    dialog_id INT REFERENCES dialogs(id) ON DELETE CASCADE,
    username TEXT REFERENCES users(username) ON DELETE CASCADE,
    PRIMARY KEY (dialog_id, username)
);