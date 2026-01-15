-- name: CreateCategory :one
insert into category (name, point_amount, parent_category, comment)
values ($1, $2, $3, $4)
returning uuid;

-- name: GetCategoryByUUID :one
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where c.uuid = $1;

-- name: ListCategoriesWithPagination :many
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value, count(c.uuid) over() as total_amount
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
limit $1 offset $2;

-- name: ListCategories :many
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status';

-- name: GetCategoryByNameAndParentNull :one
select *
from category
where name = $1 and parent_category is null;

-- name: GetCategoryByNameAndParentUUID :one
select *
from category
where name = $1 and parent_category = $2;

-- name: DeleteCategory :exec
delete
from category
where uuid = $1;

-- name: UpdateCategory :exec
update category
set name         = $1,
    point_amount = $2,
    comment = $3,
    status_uuid  = (select s.uuid from status s where s.display_value = $4 and s.type = 'category_status')
where category.uuid = $5;