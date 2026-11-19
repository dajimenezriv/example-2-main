BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

-- Separate transaction: VALIDATE only takes SHARE UPDATE EXCLUSIVE, writes keep flowing.
ALTER TABLE accounts VALIDATE CONSTRAINT accounts_actor_id_fkey;

COMMIT;
