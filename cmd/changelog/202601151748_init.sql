-- +goose Up
-- +goose StatementBegin

alter table category
    drop column status_uuid;
alter table category
    add column status_uuid uuid;

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

        UPDATE category
        SET status_uuid = default_uuid
        WHERE status_uuid IS NULL;
    END
$$;

alter table category alter column status_uuid set not null;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table category
    drop column status_uuid;
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

        UPDATE category
        SET status_uuid = default_uuid
        WHERE status_uuid IS NULL;
    END
$$;

alter table category
    alter column status_uuid set not null;

-- +goose StatementEnd
