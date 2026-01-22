-- +goose Up
-- +goose StatementBegin

alter table acheivement_category rename to achievement_category;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table achievement_category rename to acheivement_category;

-- +goose StatementEnd
