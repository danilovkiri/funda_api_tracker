-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE SCHEMA listings;

CREATE TABLE listings.listings
(
    name                TEXT            NOT NULL,
    url                 TEXT            NOT NULL,
    description         TEXT            NOT NULL,
    address_street      TEXT            NOT NULL,
    address_locality    TEXT            NOT NULL,
    address_region      TEXT            NOT NULL,
    currency            TEXT            NOT NULL,
    price               NUMERIC         NOT NULL,
    last_seen		    TIMESTAMPTZ     NOT NULL
);

CREATE TABLE listings.search_query
(
    search_query        TEXT            NOT NULL,
    updated_at		    TIMESTAMPTZ     NOT NULL
);


CREATE UNIQUE INDEX ON listings.listings (url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP SCHEMA listings CASCADE;
-- +goose StatementEnd
