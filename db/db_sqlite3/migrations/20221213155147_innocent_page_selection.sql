
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages ADD COLUMN innocent_page_id integer;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

