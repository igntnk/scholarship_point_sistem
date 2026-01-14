-- +goose Up
-- +goose StatementBegin

alter table category add constraint category_name_unique unique (name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table category drop constraint category_name_unique;

-- +goose StatementEnd
