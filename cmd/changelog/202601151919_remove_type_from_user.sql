-- +goose Up
-- +goose StatementBegin

alter table sys_user drop column type;
drop type system_user_type;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

create type system_user_type as enum ('student', 'admin');
alter table sys_user add column type system_user_type;

-- +goose StatementEnd
