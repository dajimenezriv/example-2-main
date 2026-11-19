BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

ALTER TABLE accounts
ALTER COLUMN version
SET DEFAULT 1;

ALTER TABLE actors
ALTER COLUMN version
SET DEFAULT 1;

COMMIT;
