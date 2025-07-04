-- +goose Up
CREATE TABLE note (
	created TIMESTAMP WITHOUT TIME ZONE,
	deleted TIMESTAMP WITHOUT TIME ZONE,
	latitude FLOAT,
	longitude FLOAT,
	text TEXT NOT NULL,
	updated TIMESTAMP WITHOUT TIME ZONE,

	uuid TEXT,
	PRIMARY KEY(uuid)
);

CREATE TABLE history_note (
	created TIMESTAMP WITHOUT TIME ZONE,
	latitude FLOAT,
	longitude FLOAT,
	text TEXT NOT NULL,

	version INT,
	uuid TEXT,
	PRIMARY KEY(uuid, version)
);

CREATE TABLE note_audio_recording (
	created TIMESTAMP WITHOUT TIME ZONE,
	deleted TIMESTAMP WITHOUT TIME ZONE,
	duration INTERVAL NOT NULL,
	note_uuid TEXT REFERENCES note(uuid),
	transcript TEXT,

	uuid TEXT,
	PRIMARY KEY(uuid)
);

CREATE TABLE history_note_audio_recording (
	created TIMESTAMP WITHOUT TIME ZONE,
	duration INTERVAL NOT NULL,
	note_uuid TEXT REFERENCES note(uuid),
	transcript TEXT,

	uuid TEXT,
	version INT,
	PRIMARY KEY(uuid, version)
);

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

-- +goose Down
DROP TABLE note;
DROP TABLE history_note;
DROP TABLE note_audio_recording;
DROP TABLE history_note_audio_recording;
DROP TABLE note_image;
DROP TABLE history_note_image;
