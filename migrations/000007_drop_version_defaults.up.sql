BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

-- The domain always sets version on insert; a DB default would hide a missing value.
ALTER TABLE actors
ALTER COLUMN version
DROP DEFAULT;

ALTER TABLE accounts
ALTER COLUMN version
DROP DEFAULT;

COMMIT;
