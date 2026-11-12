SET lock_timeout = '1s';

SET statement_timeout = '10min';

DROP INDEX CONCURRENTLY IF EXISTS actors_tax_id_key;

RESET statement_timeout;

RESET lock_timeout;
