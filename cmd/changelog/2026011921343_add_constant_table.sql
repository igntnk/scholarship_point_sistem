-- +goose Up
-- +goose StatementBegin

create table constants (
    name varchar unique not null ,
    value varchar not null
) ;

insert into constants (name, value) VALUES ('available_student_grades', 10);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table constants;

-- +goose StatementEnd
