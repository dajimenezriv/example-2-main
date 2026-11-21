-- +goose Up
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE actors
ALTER COLUMN version
DROP DEFAULT;

ALTER TABLE accounts
ALTER COLUMN version
DROP DEFAULT;

-- +goose Down
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE accounts
ALTER COLUMN version
SET DEFAULT 1;

ALTER TABLE actors
ALTER COLUMN version
SET DEFAULT 1;
