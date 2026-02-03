-- +goose Up
-- +goose StatementBegin

insert into status (internal_value, display_value, type) VALUES ('unactive', 'Неактивна', 'category_value_status');
insert into status (internal_value, display_value, type) VALUES ('active', 'Активна', 'category_value_status');

alter table category_value
    add column status_uuid uuid references status(uuid);

DO
$$
    DECLARE
        default_uuid UUID;
    BEGIN
        SELECT uuid INTO default_uuid
        FROM status
        WHERE internal_value = 'active'
          AND type = 'category_value_status';

        EXECUTE format('ALTER TABLE category_value ALTER COLUMN status_uuid SET DEFAULT %L', default_uuid);

        update category_value set status_uuid = default_uuid where uuid is not null;

    END
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

delete from status where type = 'category_value_status';
alter table category_value drop column status_uuid;

-- +goose StatementEnd
