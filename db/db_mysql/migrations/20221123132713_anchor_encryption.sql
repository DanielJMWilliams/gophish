
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages ADD COLUMN anchor_encryption BOOLEAN;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

