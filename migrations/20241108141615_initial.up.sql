BEGIN;
CREATE EXTENSION citext;
CREATE TABLE account
(
    id bigserial PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    email_confirmed boolean NOT NULL,
    password_hash varchar(64) NOT NULL,
    created_at timestamp NOT NULL
);
COMMIT;