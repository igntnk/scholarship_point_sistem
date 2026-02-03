-- name: CreateCategory :one
insert into category (name, point_amount, parent_category, status_uuid)
values ($1, $2, $3, (select s.uuid from status s where type = 'category_status' and internal_value = 'active'))
returning uuid;

-- name: CreateCategoryValues :batchexec
insert into category_value (name, category_uuid, point)
VALUES ($1, $2, $3)
on conflict (name, category_uuid) do update set point = $3, status_uuid = (select s.uuid from status s where type = 'category_status' and internal_value = 'active');

-- name: DeleteCategoryValues :exec
delete from category_value where category_uuid = $1;

-- name: GetCategoryByUUID :one
select c.uuid, c.name, c.point_amount, c.parent_category, c.comment, s.display_value
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
where c.uuid = $1;

-- name: ListParentCategoriesWithPagination :many
select c.uuid, c.name, c.point_amount, s.display_value, count(c.uuid) over () as total_amount, count(c_c.uuid) as sub_amount
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
         left join category c_c on c_c.parent_category = c.uuid and c_c.status_uuid != (select i_s.uuid from status i_s where i_s.type = 'category_status' and i_s.internal_value = 'unactive')
where c.parent_category is null
and s.internal_value != 'unactive'
group by c.name, c.uuid, c.point_amount, c.comment, s.display_value
limit $1 offset $2;

-- name: ListParentCategories :many
select c.uuid, c.name, c.point_amount, c.comment, s.display_value, count(c_c.uuid) as sub_amount
from category c
         join status s on c.status_uuid = s.uuid and type = 'category_status'
         left join category c_c on c_c.parent_category = c.uuid and c_c.status_uuid != (select i_s.uuid from status i_s where i_s.type = 'category_status' and i_s.internal_value = 'unactive')
where c.parent_category is null
  and s.internal_value != 'unactive'
group by c.name, c.uuid, c.point_amount, c.comment, s.display_value;

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
         left join category_value cv on c.uuid = cv.category_uuid and cv.status_uuid != (select s.uuid from status s where type = 'category_value_status' and internal_value = 'unactive')
         join status s on c.status_uuid = s.uuid
where c.parent_category = $1
  and s.internal_value != 'unactive';

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
where achievement_uuid = $1
  and s.internal_value != 'unactive';

-- name: GetParentCategoryByAchievement :one
select c.*, s.display_value as status_value
from category c
         join achievement_category ac on c.uuid = ac.category_uuid
         join status s on s.uuid = c.status_uuid
where achievement_uuid = $1
  and c.parent_category is null
  and s.internal_value != 'unactive';

-- name: GetCategoryValuesBySubCategoryUUID :many
select cv.uuid, s.display_value as status, cv.point, cv.name from category_value cv
         join status s on s.uuid = cv.status_uuid
         where cv.category_uuid = $1;

-- name: MakeCategoryValueUnactive :batchexec
update category_value cv
set status_uuid = (select s.uuid from status s where s.type = 'category_value_status' and s.internal_value = 'unactive')
where cv.name = $1;

-- name: RecreateCategoryValues :batchexec
update category_value cv
set status_uuid = (select s.uuid from status s where s.type = 'category_value_status' and s.internal_value = 'active'),
    point = $1
where cv.name = $2;
