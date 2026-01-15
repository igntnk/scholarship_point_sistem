-- +goose Up
-- +goose StatementBegin

insert into status (internal_value, display_value, type) values ('disabled', 'Отключен', 'user_status');
alter table status add constraint internal_value_type_unique unique (internal_value, type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

delete from status where internal_value = 'disabled' and type = 'user_status';
alter table status drop constraint internal_value_type_unique;

-- +goose StatementEnd
