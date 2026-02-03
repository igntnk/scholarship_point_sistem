-- +goose Up
-- +goose StatementBegin

with user_uuid as (select uuid
                   from auth_group
                   where name = 'Пользователи')
insert
into group_resource (group_uuid, resource_uuid)
VALUES ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /user/simple/:var')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /category/:var')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /user/simple')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'POST - /achievement')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /achievement/by_token')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'PUT - /achievement')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /category/parent')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /user/me')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'PUT - /user/me')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'POST - /rating')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /constant/grades_amount')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /category/children/:var')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /achievement/:var')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'GET - /rating/short_info')),
       ((select uuid from user_uuid),
        (select uuid from resource where value = 'DELETE - /achievement/:var'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

delete from group_resource where group_uuid = (select uuid from auth_group where name = 'Пользователи');

-- +goose StatementEnd
