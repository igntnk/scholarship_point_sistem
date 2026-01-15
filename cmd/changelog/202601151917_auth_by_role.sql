-- +goose Up
-- +goose StatementBegin

drop table user_resource;

create table auth_role
(
    uuid uuid primary key default uuid_generate_v4(),
    name varchar unique not null
);

create table auth_group
(
    uuid uuid primary key default uuid_generate_v4(),
    name varchar unique not null
);

create table user_role
(
    user_uuid uuid references sys_user (uuid)  not null,
    role_uuid uuid references auth_role (uuid) not null
);
alter table user_role
    add constraint user_role_unique unique (user_uuid, role_uuid);

create table role_group
(
    role_uuid  uuid references auth_role (uuid)  not null,
    group_uuid uuid references auth_group (uuid) not null
);
alter table role_group
    add constraint role_group_unique unique (role_uuid, group_uuid);

create table group_resource
(
    group_uuid    uuid references auth_group (uuid) not null,
    resource_uuid uuid references resource (uuid)   not null
);
alter table group_resource
    add constraint group_resource_unique unique (group_uuid, resource_uuid);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

create table resource (
                          uuid uuid primary key default uuid_generate_v4(),
                          value varchar
);

drop table auth_role;
drop table auth_group;
drop table user_role;
drop table role_group;
drop table group_resource;

-- +goose StatementEnd
