-- +goose Up
-- +goose StatementBegin

update status set type = 'category_status' where type = 'category_staus'

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

update status set type = 'category_staus' where type = 'category_status'

-- +goose StatementEnd
