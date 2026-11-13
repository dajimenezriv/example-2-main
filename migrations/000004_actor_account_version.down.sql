BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

ALTER TABLE accounts
DROP COLUMN version;

ALTER TABLE actors
DROP COLUMN version;

COMMIT;
