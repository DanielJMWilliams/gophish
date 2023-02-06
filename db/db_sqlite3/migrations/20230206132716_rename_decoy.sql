
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages RENAME COLUMN innocent_page_id TO decoy_page_id;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

