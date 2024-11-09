BEGIN;
CREATE EXTENSION citext;
CREATE TABLE account
(
    id bigserial PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    email_confirmed boolean NOT NULL,
    password_hash VARCHAR(512) NOT NULL,
    password_salt VARCHAR(512) NOT NULL
);
COMMIT;