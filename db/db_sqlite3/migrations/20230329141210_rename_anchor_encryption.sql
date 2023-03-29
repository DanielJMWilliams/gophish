
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages RENAME COLUMN anchor_encryption TO proxy_bypass_enabled;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

