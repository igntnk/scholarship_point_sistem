-- name: CreateUser :one
insert into sys_user(name, second_name, patronymic, gradebook_number, birth_date, email, phone_number, password, salt)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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

-- name: GetUserByEmail :one
select * from sys_user where email = $1;

-- name: MakeUserVerified :exec
update sys_user set status_uuid = (select s.uuid from status s where type = 'user_status' and internal_value = 'approved');

-- name: MakeUserUnverified :exec
update sys_user set status_uuid = (select s.uuid from status s where type = 'user_status' and internal_value = 'declined');

-- name: GetShortRatingInfo :many
with sub as (select distinct row_number() over ()           as position,
                             u.uuid,
                             u.name,
                             second_name,
                             patronymic,
                             gradebook_number,
                             birth_date,
                             phone_number,
                             email,
                             u_s.display_value              as user_status,
                             sum(cv.point) + c.point_amount as point_amount,
                             count(a.uuid)                  as achievement_amount,
                             count(*) over ()               as total_amount,
                             case
                                 when max(case when a_s.internal_value = 'unapproved' then 1 else 0 end) = 1
                                     then false
                                 else true
                                 end                        as all_achievement_verified
             from sys_user u
                      left join achievement a on a.user_uuid = u.uuid and a.status_uuid in (select s.uuid
                                                                                            from status s
                                                                                            where s.type = 'achievement_status'
                                                                                              and s.internal_value != 'declined')
                      left join achievement_category ac on ac.achievement_uuid = a.uuid
                      left join category c
                                on c.uuid = ac.category_uuid and parent_category is null and
                                   c.status_uuid in (select s.uuid
                                                     from status s
                                                     where s.type = 'category_status'
                                                       and s.internal_value = 'active')
                      left join category c_p on c_p.uuid = ac.category_uuid and c_p.parent_category is not null and
                                                c_p.status_uuid in (select s.uuid
                                                                    from status s
                                                                    where s.type = 'category_status'
                                                                      and s.internal_value = 'active')
                      left join achievement_category_value acv on acv.achievement_uuid = a.uuid
                      left join category_value cv on cv.uuid = acv.category_value_uuid
                      left join status u_s on u_s.uuid = u.status_uuid and u_s.type = 'user_status'
                      left join status a_s on a_s.uuid = a.status_uuid and a_s.type = 'achievement_status'
             where c.point_amount is not null
             group by u.uuid, u.name, second_name, patronymic, gradebook_number, birth_date, phone_number, email,
                      c.point_amount, u_s.display_value

             order by c.point_amount)
select s.position,
       s.uuid,
       s.name,
       s.second_name,
       s.patronymic,
       s.gradebook_number,
       s.birth_date,
       s.phone_number,
       s.email,
       s.user_status,
       s.point_amount,
       s.achievement_amount,
       s.total_amount,
       s.all_achievement_verified
from sub s
where s.uuid = $1
   or s.position = 1
order by position;