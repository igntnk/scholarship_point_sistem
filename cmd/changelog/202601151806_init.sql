-- +goose Up
-- +goose StatementBegin

alter table category drop constraint category_name_unique;
alter table category add constraint category_name_par_uuid_unique unique (name, parent_category);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table category drop constraint category_name_par_uuid_unique;
alter table category add constraint category_name_unique unique (name);

-- +goose StatementEnd
