BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

DROP TRIGGER accounts_audit_delete ON accounts;

DROP TRIGGER actors_audit_delete ON actors;

DROP FUNCTION audit_delete ();

DROP TABLE deletion_audit;

ALTER TABLE accounts
DROP CONSTRAINT accounts_actor_id_fkey;

ALTER TABLE accounts
ADD CONSTRAINT accounts_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actors (id) NOT VALID;

COMMIT;

BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

ALTER TABLE accounts VALIDATE CONSTRAINT accounts_actor_id_fkey;

COMMIT;
