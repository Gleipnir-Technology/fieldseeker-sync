-- +goose Up
ALTER TABLE note_audio ADD COLUMN is_audio_normalized BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ALTER COLUMN is_audio_normalized DROP DEFAULT;
-- +goose Down
ALTER TABLE note_audio DROP COLUMN is_audio_normalized;
