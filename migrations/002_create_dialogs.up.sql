CREATE TABLE dialogs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE dialog_members (
    dialog_id INT NOT NULL REFERENCES dialogs(id) ON DELETE CASCADE,
    username TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    PRIMARY KEY (dialog_id, username)
);

CREATE INDEX idx_dialog_members_username ON dialog_members(username);