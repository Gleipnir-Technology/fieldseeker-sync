-- +goose Up
ALTER TABLE note_audio DROP COLUMN has_been_reviewed;
ALTER TABLE note_audio DROP COLUMN transcription_internally_edited;
ALTER TABLE note_audio DROP COLUMN needs_further_review;
-- +goose Down
ALTER TABLE note_audio ADD COLUMN has_been_reviewed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ADD COLUMN transcription_internally_edited BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE note_audio ADD COLUMN needs_further_review BOOLEAN NOT NULL DEFAULT false;
