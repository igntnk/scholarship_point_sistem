-- +goose Up
-- +goose StatementBegin

create extension if not exists "uuid-ossp";

create table if not exists status
(
    uuid           uuid primary key default uuid_generate_v4(),
    internal_value varchar(100),
    display_value  varchar(100),
    type           varchar(40)
    );

insert into status (internal_value, display_value, type)
values ('unapproved', 'Неподтвержденный', 'user_status');
insert into status (internal_value, display_value, type)
values ('approved', 'Подтвержденный', 'user_status');
insert into status (internal_value, display_value, type)
values ('used', 'Отклонен', 'user_status');

insert into status (internal_value, display_value, type)
values ('unapproved', 'Неподтвержденный', 'achievement_status');
insert into status (internal_value, display_value, type)
values ('approved', 'Подтвержденный', 'achievement_status');
insert into status (internal_value, display_value, type)
values ('declined', 'Отклонен', 'achievement_status');
insert into status (internal_value, display_value, type)
values ('used', 'Использован', 'achievement_status');

create table if not exists category
(
    uuid            uuid primary key default uuid_generate_v4(),
    name            varchar,
    point_amount    numeric(5, 2),
    parent_category uuid references category (uuid),
    comment         varchar
    );

create type system_user_type as enum ('student', 'admin');

create table if not exists sys_user(
                                       uuid             uuid primary key default uuid_generate_v4(),
    name             varchar(30)  not null,
    second_name      varchar(40)  not null,
    patronymic       varchar(40),
    gradebook_number varchar(100) not null,
    birth_date       date,
    email            varchar,
    phone_number     varchar(20),
    status_uuid      uuid references status (uuid),
    type             system_user_type
    );

insert into sys_user (name, second_name, gradebook_number, type)
values ('Админ', 'Главный', '0000-0000', 'admin');

create table if not exists achievement
(
    uuid            uuid primary key default uuid_generate_v4(),
    comment         text,
    attachment_link text                               not null,
    user_uuid       uuid references sys_user (uuid) not null,
    status_uuid     uuid references status (uuid)      not null
    );

create table if not exists acheivement_category
(
    category_uuid    uuid references category (uuid),
    achievement_uuid uuid references achievement (uuid)
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS acheivement_category;
DROP TABLE IF EXISTS achievement;
DROP TABLE IF EXISTS sys_user;
DROP TYPE IF EXISTS system_user_type;
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS status;

DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
