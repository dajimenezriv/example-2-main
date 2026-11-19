-- +goose Up
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';

-- Separate transaction: VALIDATE only takes SHARE UPDATE EXCLUSIVE, writes keep flowing.
ALTER TABLE accounts VALIDATE CONSTRAINT accounts_actor_id_fkey;

-- +goose Down
-- Nothing to undo: an already-validated constraint can't be un-validated.
