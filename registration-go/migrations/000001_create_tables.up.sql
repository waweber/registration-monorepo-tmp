BEGIN;
CREATE TABLE IF NOT EXISTS registration (
    id VARCHAR(300) NOT NULL,
    event_id VARCHAR(300) NOT NULL,
    status VARCHAR(16) NOT NULL,
    version INTEGER NOT NULL,
    date_created TIMESTAMP WITH TIME ZONE NOT NULL,
    date_updated TIMESTAMP WITH TIME ZONE,
    options JSONB,
    first_name VARCHAR(300),
    last_name VARCHAR(300),
    preferred_name VARCHAR(300),
    nickname VARCHAR(300),
    number INTEGER,
    email VARCHAR(300),
    account_id VARCHAR(300),
    checked_in BOOLEAN,
    date_checked_in TIMESTAMP WITH TIME ZONE,
    extra_data JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS registration_options ON registration USING GIN (options jsonb_path_ops);
CREATE INDEX IF NOT EXISTS registration_extra_data ON registration USING GIN (extra_data jsonb_path_ops);
COMMIT;
