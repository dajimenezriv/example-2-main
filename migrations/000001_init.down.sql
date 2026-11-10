BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

DROP TABLE accounts;

DROP TABLE actors;

COMMIT;