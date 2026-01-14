-- name: CreateCategory :one
insert into category (name, point_amount, parent_category, comment)
values ($1, $2, $3, $4)
    returning uuid;

-- name: GetCategoryByUUID :one
select * from category where uuid = $1;

-- name: ListCategories :many
select * from category limit $1 offset $2;

-- name: GetCategoryByName :one
select * from category where name = $1;