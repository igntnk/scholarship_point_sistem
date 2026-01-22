-- +goose Up
-- +goose StatementBegin

insert into status (internal_value, display_value, type) values ('removed', 'Удален', 'achievement_status');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

delete from status where internal_value = 'removed' and type = 'achievement_status';

-- +goose StatementEnd
