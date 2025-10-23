-- +goose Up
ALTER TABLE note_audio_breadcrumb ADD COLUMN manually_selected BOOLEAN NOT NULL DEFAULT false;
-- +goose Down
ALTER TABLE note_audio_breadcrumb DROP COLUMN manually_selected;
