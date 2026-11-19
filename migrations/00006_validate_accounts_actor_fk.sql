-- +goose Up
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

-- VALIDATE only takes SHARE UPDATE EXCLUSIVE.
ALTER TABLE accounts VALIDATE CONSTRAINT accounts_actor_id_fkey;

-- +goose Down
