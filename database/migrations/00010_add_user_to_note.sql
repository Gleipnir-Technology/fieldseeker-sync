-- +goose Up
ALTER TABLE note_audio ADD COLUMN creator INTEGER REFERENCES user_(id);
UPDATE note_audio SET creator = 1;
ALTER TABLE note_audio ALTER COLUMN creator SET NOT NULL;

ALTER TABLE note_image ADD COLUMN creator INTEGER REFERENCES user_(id);
UPDATE note_image SET creator = 1;
ALTER TABLE note_image ALTER COLUMN creator SET NOT NULL;


-- +goose Down
ALTER TABLE note_audio DROP COLUMN creator;
ALTER TABLE note_image DROP COLUMN creator;
