-- name: CreateCategory :one
insert into category (name, point_amount, parent_category, comment)
values ($1, $2, $3, $4)
returning uuid;

-- name: GetCategoryByUUID :one
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where c.uuid = $1;

-- name: ListParentCategoriesWithPagination :many
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value, count(c.uuid) over() as total_amount
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where parent_category is null
limit $1 offset $2;

-- name: ListParentCategories :many
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where parent_category is null;

-- name: GetCategoryByNameAndParentNull :one
select *
from category
where name = $1 and parent_category is null;

-- name: GetCategoryByNameAndParentUUID :one
select *
from category
where name = $1 and parent_category = $2;

-- name: GetChildCategories :many
select * from category
where parent_category = $1;

-- name: DeleteCategory :exec
update category c set status_uuid = (select s.uuid from status s where s.internal_value = 'unactive' and s.type = 'category_status')
where c.uuid = $1;

-- name: UpdateCategory :exec
update category
set name         = $1,
    point_amount = $2,
    comment = $3,
    status_uuid  = (select s.uuid from status s where s.display_value = $4 and s.type = 'category_status')
where category.uuid = $5;

-- name: GetCategoryByAchievement :many
select c.*, s.display_value as status_value
from category c
         join achievement_category ac on c.uuid = ac.category_uuid
         join status s on s.uuid = c.status_uuid
where achievement_uuid = $1;