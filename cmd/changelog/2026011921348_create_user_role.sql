-- +goose Up
-- +goose StatementBegin

insert into auth_role (name)
VALUES ('Пользователи');
insert into auth_group (name)
values ('Пользователи');
insert into role_group (role_uuid, group_uuid)
VALUES (
        (select uuid from auth_role where name = 'Пользователи'),
        (select uuid from auth_group where name = 'Пользователи')
       );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

delete from role_group where role_uuid = (select uuid from auth_role where name = 'Пользователи');
delete from auth_role where name = 'Пользователи';
delete from auth_group where name = 'Пользователи';

-- +goose StatementEnd
