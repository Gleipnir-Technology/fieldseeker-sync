-- +goose Up
ALTER TABLE note_audio ADD COLUMN needs_further_review BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ALTER COLUMN needs_further_review DROP DEFAULT;
-- +goose Down
ALTER TABLE note_audio DROP COLUMN needs_further_review;
