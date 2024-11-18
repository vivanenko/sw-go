CREATE TABLE refresh_token
(
    id serial PRIMARY KEY,
    value varchar(64) NOT NULL UNIQUE,
    expires_at timestamp NOT NULL,
    account_id bigint NOT NULL REFERENCES account (id)
);