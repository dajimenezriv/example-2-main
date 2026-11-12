BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

ALTER TABLE actors
ADD CONSTRAINT actors_tax_id_key UNIQUE USING INDEX actors_tax_id_key;

COMMIT;
