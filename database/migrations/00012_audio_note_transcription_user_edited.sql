-- +goose Up
ALTER TABLE note_audio ADD COLUMN transcription_user_edited BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ALTER COLUMN transcription_user_edited DROP DEFAULT;
-- +goose Down
ALTER TABLE note_audio DROP COLUMN transcription_user_edited;
