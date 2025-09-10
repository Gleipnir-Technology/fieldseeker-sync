-- +goose Up
ALTER TABLE note_audio ADD COLUMN is_transcoded_to_ogg BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ALTER COLUMN is_transcoded_to_ogg DROP DEFAULT;
-- +goose Down
ALTER TABLE note_audio DROP COLUMN is_transcoded_to_ogg;
