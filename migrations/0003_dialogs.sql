CREATE TABLE IF NOT EXISTS dialogs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dialog_users (
    dialog_id BIGINT NOT NULL REFERENCES dialogs(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (dialog_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_dialog_users_user_id ON dialog_users(user_id);
