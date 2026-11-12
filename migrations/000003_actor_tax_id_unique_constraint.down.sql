BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

-- Dropping the constraint also drops its backing index.
ALTER TABLE actors
DROP CONSTRAINT actors_tax_id_key;

COMMIT;
