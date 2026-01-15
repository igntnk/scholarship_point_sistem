-- +goose Up
-- +goose StatementBegin

alter table sys_user add column password varchar;

create table resource (
                          uuid uuid primary key default uuid_generate_v4(),
                          value varchar
);

create table user_resource (
                               resource_uuid uuid not null ,
                               user_uuid uuid not null
);

alter table user_resource add constraint user_resource_unique unique(resource_uuid, user_uuid);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table sys_user drop column password;
drop table resource;
drop table user_resource;

-- +goose StatementEnd
