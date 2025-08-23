-- +goose Up
DROP TABLE note_image;
DROP TABLE history_note_image;
CREATE TABLE note_image (
	created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
	deleted TIMESTAMP WITHOUT TIME ZONE,

	version INT,
	uuid TEXT,

	PRIMARY KEY(version, uuid)
);
-- +goose Down
DROP TABLE note_image;

CREATE TABLE note_image (
	created TIMESTAMP WITHOUT TIME ZONE,
	deleted TIMESTAMP WITHOUT TIME ZONE,
	size_x INT NOT NULL,
	size_y INT NOT NULL,
	note_uuid TEXT REFERENCES note(uuid),

	uuid TEXT,
	PRIMARY KEY(uuid)
);

CREATE TABLE history_note_image (
	created TIMESTAMP WITHOUT TIME ZONE,
	size_x INT NOT NULL,
	size_y INT NOT NULL,
	note_uuid TEXT REFERENCES note(uuid),

	uuid TEXT,
	version INT,
	PRIMARY KEY(uuid, version)
);

