-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE listings ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT '0001-01-01 00:00:00';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE listings DROP column created_at;
-- +goose StatementEnd
