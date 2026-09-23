-- +goose Up
CREATE TABLE feeds (
    feed_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name       TEXT NOT NULL,
    url        TEXT NOT NULL,
    user_id    UUID NOT NULL,

    CONSTRAINT fk_users
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);

CREATE TRIGGER feeds_set_updated_at
BEFORE UPDATE ON feeds
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE feeds;