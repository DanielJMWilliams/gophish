
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE email_requests ADD COLUMN anchor varchar(32);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

