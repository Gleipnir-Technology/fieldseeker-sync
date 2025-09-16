-- +goose Up
CREATE TABLE task_audio_review (
	ID SERIAL PRIMARY KEY,
	completed_by INTEGER REFERENCES user_(id),
	created TIMESTAMP NOT NULL,
	needs_review BOOLEAN NOT NULL,
	note_audio_uuid TEXT NOT NULL,
	note_audio_version INTEGER NOT NULL,
	reviewed_by INTEGER REFERENCES user_(id),
	FOREIGN KEY (note_audio_uuid, note_audio_version) REFERENCES note_audio (uuid, version)
);

INSERT INTO task_audio_review (
    completed_by,
    created,
    needs_review,
    note_audio_uuid,
    note_audio_version,
    reviewed_by
)
SELECT 
    NULL AS completed_by,
    created,
    needs_further_review AS needs_review,
    uuid AS note_audio_uuid,
    version AS note_audio_version,
    NULL AS reviewed_by
FROM note_audio
WHERE has_been_reviewed = false;

-- +goose Down
DROP TABLE task_audio_review;
