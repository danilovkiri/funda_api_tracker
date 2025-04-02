-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE sessions
(
    user_id                     TEXT            NOT NULL,
    chat_id                     INTEGER         NOT NULL,
    update_interval_seconds     INTEGER         NOT NULL DEFAULT 3600,
    is_active                   BOOLEAN         NOT NULL DEFAULT FALSE,
    regions                     TEXT            DEFAULT '',
    cities                      TEXT            DEFAULT '',
    last_synced_at              TIMESTAMP       DEFAULT '0001-01-01 00:00:00'
);
CREATE UNIQUE INDEX sessions_unique_user_id_idx ON sessions(user_id);

CREATE TABLE search_queries
(
    user_id             TEXT            NOT NULL,
    search_query        TEXT            NOT NULL
);
CREATE UNIQUE INDEX search_queries_unique_user_id_idx ON search_queries(user_id);

CREATE TABLE listings
(
    user_id             TEXT            NOT NULL,
    name                TEXT            NOT NULL,
    url                 TEXT            NOT NULL,
    description         TEXT            NOT NULL,
    address_street      TEXT            NOT NULL,
    address_locality    TEXT            NOT NULL,
    address_region      TEXT            NOT NULL,
    currency            TEXT            NOT NULL,
    price               NUMERIC         NOT NULL,
    is_new              BOOLEAN         NOT NULL
);
CREATE UNIQUE INDEX listings_unique_user_id_url_idx ON listings(user_id, url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE listings;
DROP TABLE search_query;
-- +goose StatementEnd
