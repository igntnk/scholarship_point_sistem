-- +goose Up
-- +goose StatementBegin

drop table if exists user_achievement;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

create table user_achievement(
                                 user_uuid uuid references sys_user(uuid) not null ,
                                 achievement_uuid uuid references achievement(uuid) not null
);
alter table user_achievement add constraint user_achievement_unique unique (user_uuid, achievement_uuid);

-- +goose StatementEnd
