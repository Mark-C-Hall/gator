-- +goose Up
CREATE TABLE posts (
    post_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    title        TEXT NOT NULL,
    url          TEXT NOT NULL UNIQUE,
    description  TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    feed_id      UUID NOT NULL,

    CONSTRAINT fk_feeds
        FOREIGN KEY (feed_id)
        REFERENCES feeds(feed_id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;