-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE sessions ADD column dnd_status BOOLEAN DEFAULT FALSE;
ALTER TABLE sessions ADD column dnd_start INTEGER DEFAULT 0;
ALTER TABLE sessions ADD column dnd_end INTEGER DEFAULT 0;
ALTER TABLE sessions ADD column sync_count_since_last_change INTEGER DEFAULT 0;
ALTER TABLE listings ADD COLUMN uuid TEXT NOT NULL DEFAULT 'empty';

CREATE TABLE favorites
(
    user_id             TEXT            NOT NULL,
    name                TEXT            NOT NULL,
    url                 TEXT            NOT NULL,
    description         TEXT            NOT NULL,
    address_street      TEXT            NOT NULL,
    address_locality    TEXT            NOT NULL,
    address_region      TEXT            NOT NULL,
    currency            TEXT            NOT NULL,
    price               NUMERIC         NOT NULL
);
CREATE UNIQUE INDEX favorites_unique_user_id_url_idx ON favorites(user_id, url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE favorites;
ALTER TABLE sessions DROP column dnd_status;
ALTER TABLE sessions DROP column dnd_start;
ALTER TABLE sessions DROP column dnd_end;
ALTER TABLE sessions DROP column sync_count_since_last_change;
-- +goose StatementEnd
