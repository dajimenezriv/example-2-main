BEGIN;

SET
  LOCAL lock_timeout = '1s';

SET
  LOCAL statement_timeout = '5s';

-- Drop and re-add in one transaction so the FK is never unenforced.
-- NOT VALID skips the full scan under lock; validated in 000006.
ALTER TABLE accounts
DROP CONSTRAINT accounts_actor_id_fkey;

ALTER TABLE accounts
ADD CONSTRAINT accounts_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actors (id) ON DELETE CASCADE NOT VALID;

CREATE TABLE deletion_audit (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  entity_type TEXT NOT NULL,
  entity_id UUID NOT NULL,
  data JSONB NOT NULL,
  deleted_by TEXT,
  deleted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX deletion_audit_entity_idx ON deletion_audit (entity_type, entity_id);

-- Trigger, not app code: cascaded account deletes never reach the app.
CREATE FUNCTION audit_delete () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO deletion_audit (entity_type, entity_id, data, deleted_by)
  VALUES (TG_ARGV[0], OLD.id, to_jsonb(OLD), current_setting('app.principal', true));
  RETURN OLD;
END
$$;

CREATE TRIGGER actors_audit_delete
AFTER DELETE ON actors FOR EACH ROW
EXECUTE FUNCTION audit_delete ('actor');

CREATE TRIGGER accounts_audit_delete
AFTER DELETE ON accounts FOR EACH ROW
EXECUTE FUNCTION audit_delete ('account');

COMMIT;
