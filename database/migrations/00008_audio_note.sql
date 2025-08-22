-- +goose Up
CREATE TABLE note_audio (
	created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
	deleted TIMESTAMP WITHOUT TIME ZONE,
	duration REAL,
	transcription TEXT,

	version INT,
	uuid TEXT,

	PRIMARY KEY(version, uuid)
);

CREATE TABLE note_audio_breadcrumb (
	cell NUMERIC NOT NULL,
	created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
	note_audio_uuid TEXT NOT NULL,
	note_audio_version INT NOT NULL,
	position INT NOT NULL,

	FOREIGN KEY (note_audio_version, note_audio_uuid) REFERENCES note_audio (version, uuid),
	PRIMARY KEY (note_audio_version, note_audio_uuid, position)
);

-- +goose Down
DROP TABLE note_audio;
DROP TABLE note_audio_breadcrumb;
