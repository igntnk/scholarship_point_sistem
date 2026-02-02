-- +goose Up
-- +goose StatementBegin

create table category_value (
    uuid uuid primary key default uuid_generate_v4(),
    name varchar not null,
    category_uuid uuid references category(uuid),
    point numeric(22,3) not null
);
alter table category_value add constraint category_uuid_name_unique unique(category_uuid, name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table category_value;

-- +goose StatementEnd
