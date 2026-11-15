BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

ALTER TABLE actors
ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

ALTER TABLE accounts
ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

COMMIT;
