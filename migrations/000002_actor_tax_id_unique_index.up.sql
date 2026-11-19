-- Single statement on purpose: golang-migrate runs a multi-statement file as one implicit
-- transaction, and CONCURRENTLY can't run in one. Timeouts come from PGOPTIONS (Taskfile).
-- squawk-ignore require-timeout-settings
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS actors_tax_id_key ON actors (tax_id);
