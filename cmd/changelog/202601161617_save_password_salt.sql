-- +goose Up
-- +goose StatementBegin

alter table sys_user add column salt varchar(10);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table sys_user drop column salt;

-- +goose StatementEnd
