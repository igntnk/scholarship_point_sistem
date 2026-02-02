-- +goose Up
-- +goose StatementBegin

create table achievement_category_value
(
    achievement_uuid uuid references achievement (uuid),
    category_value_uuid    uuid references category_value(uuid)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table achievement_category_value;

-- +goose StatementEnd
