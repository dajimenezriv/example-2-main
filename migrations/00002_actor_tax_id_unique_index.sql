-- +goose NO TRANSACTION
-- squawk-ignore-file ban-concurrent-index-creation-in-transaction
-- squawk-ignore-file require-timeout-settings

-- +goose Up
-- A failed concurrent build leaves an INVALID index that IF NOT EXISTS would silently accept.
DROP INDEX CONCURRENTLY IF EXISTS actors_tax_id_key;
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS actors_tax_id_key ON actors (tax_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS actors_tax_id_key;
