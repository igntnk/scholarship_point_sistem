-- +goose Up
-- +goose StatementBegin

insert into status (internal_value, display_value, type)
values ('active', 'Активно', 'category_staus');
insert into status (internal_value, display_value, type)
values ('unactive', 'Неактивно', 'category_staus');

alter table category
    add column status_uuid varchar(100);

DO
$$
    DECLARE
        default_uuid UUID;
    BEGIN
        SELECT uuid
        INTO default_uuid
        FROM status
        WHERE internal_value = 'active'
          AND type = 'category_staus';

        EXECUTE format('ALTER TABLE category ALTER COLUMN status_uuid SET DEFAULT %L', default_uuid);
    END
$$;

alter table category
    alter column status_uuid set not null;


DO
$$
    DECLARE
        default_uuid UUID;
    BEGIN
        SELECT uuid
        INTO default_uuid
        FROM status
        WHERE internal_value = 'unapproved'
          AND type = 'user_status';

        EXECUTE format('ALTER TABLE sys_user ALTER COLUMN status_uuid SET DEFAULT %L', default_uuid);

        UPDATE sys_user
        SET status_uuid = default_uuid
        WHERE status_uuid IS NULL;
    END
$$;

alter table sys_user
    alter column status_uuid set not null;

DO
$$
    DECLARE
        default_uuid UUID;
    BEGIN
        SELECT uuid
        INTO default_uuid
        FROM status
        WHERE internal_value = 'unapproved'
          AND type = 'achievement_status';

        EXECUTE format('ALTER TABLE achievement ALTER COLUMN status_uuid SET DEFAULT %L', default_uuid);

    END
$$;



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE achievement
    ALTER COLUMN status_uuid DROP NOT NULL,
    ALTER
        COLUMN status_uuid DROP
        DEFAULT;

ALTER TABLE sys_user
    ALTER COLUMN status_uuid DROP NOT NULL,
    ALTER
        COLUMN status_uuid DROP
        DEFAULT;

ALTER TABLE category
    DROP
        COLUMN status_uuid;

DELETE
FROM status
WHERE type = 'category_staus';


-- +goose StatementEnd
