BEGIN;
DROP TABLE IF EXISTS access_code;
DROP TABLE IF EXISTS event_stats;
DROP INDEX IF EXISTS registration_extra_data;
DROP INDEX IF EXISTS registration_options;
DROP TABLE IF EXISTS registration;
COMMIT;
