-- +goose Up
ALTER TABLE note_audio ADD COLUMN has_been_reviewed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ALTER COLUMN has_been_reviewed DROP DEFAULT;
-- +goose Down
ALTER TABLE note_audio DROP COLUMN has_been_reviewed;
