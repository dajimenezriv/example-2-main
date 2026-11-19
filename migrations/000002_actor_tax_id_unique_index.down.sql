-- Single statement on purpose: see the up migration.
-- squawk-ignore require-timeout-settings
DROP INDEX CONCURRENTLY IF EXISTS actors_tax_id_key;
