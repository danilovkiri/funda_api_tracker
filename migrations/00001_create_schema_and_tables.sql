-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE listings
(
    name                TEXT            NOT NULL,
    url                 TEXT            NOT NULL,
    description         TEXT            NOT NULL,
    address_street      TEXT            NOT NULL,
    address_locality    TEXT            NOT NULL,
    address_region      TEXT            NOT NULL,
    currency            TEXT            NOT NULL,
    price               NUMERIC         NOT NULL,
    last_seen		    DATETIME        NOT NULL
);

CREATE TABLE search_query
(
    search_query        TEXT            NOT NULL,
    updated_at		    DATETIME        NOT NULL
);


CREATE UNIQUE INDEX listings_url_unique_idx ON listings(url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE listings;
DROP TABLE search_query;
-- +goose StatementEnd
