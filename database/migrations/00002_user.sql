-- +goose Up
CREATE TYPE HashType AS ENUM ('bcrypt-14');

CREATE TABLE user_ (
	ID SERIAL PRIMARY KEY,
	display_name VARCHAR(200),
	password_hash_type HashType,
	password_hash TEXT,
	username TEXT
);

-- +goose Down
DROP TABLE user_;
DROP TYPE HashType;
