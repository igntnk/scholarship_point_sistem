-- +goose Up
-- +goose StatementBegin

alter table category alter column name set not null;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table category alter column name drop not null;

-- +goose StatementEnd
