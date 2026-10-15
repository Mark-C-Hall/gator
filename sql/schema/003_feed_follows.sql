-- +goose Up
CREATE TABLE feed_follows (
    feed_follow_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id        UUID NOT NULL,
    feed_id        UUID NOT NULL,

    CONSTRAINT fk_users
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_feeds
        FOREIGN KEY (feed_id)
        REFERENCES feeds(feed_id)
        ON DELETE CASCADE,

    UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;