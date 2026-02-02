-- name: CreateCategory :one
insert into category (name, point_amount, parent_category)
values ($1, $2, $3)
returning uuid;

-- name: CreateCategoryValues :batchexec
insert into category_value (name, category_uuid, point)
VALUES ($1, $2, $3);

-- name: DeleteCategoryValues :exec
delete from category_value where category_uuid = $1;

-- name: GetCategoryByUUID :one
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where c.uuid = $1;

-- name: ListParentCategoriesWithPagination :many
select c.uuid, c.name, c.point_amount, s.display_value, count(c.uuid) over () as total_amount
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where parent_category is null
limit $1 offset $2;

-- name: ListParentCategories :many
select c.uuid, c.name, c.point_amount, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where parent_category is null;

-- name: GetCategoryByNameAndParentNull :one
select *
from category
where name = $1
  and parent_category is null;

-- name: GetCategoryByNameAndParentUUID :one
select *
from category
where name = $1
  and parent_category = $2;

-- name: GetChildCategories :many
select c.uuid, c.name as category_name,cv.name as value_name, cv.point
from category c
         join category_value cv on c.uuid = cv.category_uuid
where c.parent_category = $1;

-- name: DeleteCategory :exec
update category c
set status_uuid = (select s.uuid from status s where s.internal_value = 'unactive' and s.type = 'category_status')
where c.uuid = $1;

-- name: UpdateCategory :exec
update category
set name         = $1,
    point_amount = $2,
    status_uuid  = (select s.uuid from status s where s.display_value = $3 and s.type = 'category_status')
where category.uuid = $4;

-- name: GetCategoryByAchievement :many
select c.*, s.display_value as status_value
from category c
         join achievement_category ac on c.uuid = ac.category_uuid
         join status s on s.uuid = c.status_uuid
where achievement_uuid = $1;