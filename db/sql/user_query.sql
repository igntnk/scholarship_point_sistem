-- name: CreateUser :one
insert into sys_user(name, second_name, patronymic, gradebook_number, birth_date, email, phone_number, password)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning uuid;

-- name: GetSimpleUserByUUID :one
select *
from sys_user
where uuid = $1;

-- name: GetSimpleUserList :many
select *
from sys_user;

-- name: GetSimpleUserListWithPagination :many
select *, count(*) over() as total_amount
from sys_user
limit $1 offset $2;

-- name: GetApprovedUserByGradeBookNumber :one
select * from sys_user u
                  join status s on u.status_uuid = s.uuid and s.type = 'user_status'
where u.gradebook_number = $1 and s.internal_value = 'approved';

-- name: UpdateUserInfoWithoutGradeBook :exec
update sys_user
set name         = $1,
    second_name  = $2,
    patronymic   = $3,
    birth_date   = $4,
    phone_number = $5,
    email        = $6
where uuid = $7;

-- name: UpdateUserInfoWithGradeBook :exec
update sys_user
set name             = $1,
    second_name      = $2,
    patronymic       = $3,

    birth_date       = $4,
    phone_number     = $5,
    email            = $6,
    gradebook_number = $7,
    status_uuid      = (select status.uuid from status where internal_value = 'unapproved' and type = 'user_status')
where sys_user.uuid = $8;