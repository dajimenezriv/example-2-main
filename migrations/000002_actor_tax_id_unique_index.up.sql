SET lock_timeout = '1s';

-- A concurrent build doesn't block writes, so allow it time on large tables.
SET statement_timeout = '10min';

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS actors_tax_id_key ON actors (tax_id);

RESET statement_timeout;

RESET lock_timeout;
