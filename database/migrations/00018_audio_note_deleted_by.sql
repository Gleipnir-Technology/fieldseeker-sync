-- +goose Up
ALTER TABLE note_audio ADD COLUMN deleted_by INT REFERENCES user_;
-- +goose Down
ALTER TABLE note_audio DROP COLUMN deleted_by;
