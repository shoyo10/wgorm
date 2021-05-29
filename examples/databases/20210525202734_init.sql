-- +goose Up
CREATE TABLE IF NOT EXISTS users (
	id bigserial,
	email varchar(128) NOT NULL,
	password varchar(64) NOT NULL,
	age int,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	deleted_at timestamp,
	PRIMARY KEY (id),
	CONSTRAINT uniq_email UNIQUE (email)
);

-- +goose Down
DROP TABLE IF EXISTS users;
