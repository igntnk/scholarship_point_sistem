-- +goose Up
-- +goose StatementBegin

alter table achievement add column achievement_date date not null default now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table achievement drop column achievement_date;
-- +goose StatementEnd
