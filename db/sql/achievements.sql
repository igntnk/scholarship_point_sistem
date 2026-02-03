-- name: GetSimpleUserAchievementByUUID :one
select a.*,
       c.name                         as category_name,
       s.display_value                as status,
       c.uuid                         as category_uuid,
       c.point_amount                 as base_point_amount,
       coalesce(sum(cv.point) + c.point_amount, 0)::numeric as point_amount,
       count(*) over ()               as total_records
from achievement a
         join achievement_category ac on ac.achievement_uuid = a.uuid
         left join category c on c.uuid = ac.category_uuid and parent_category is null
         left join category c_p on c_p.uuid = ac.category_uuid and c_p.parent_category is not null
         join achievement_category_value acv on acv.achievement_uuid = a.uuid
         join category_value cv on cv.uuid = acv.category_value_uuid and cv.category_uuid = c.uuid
         join status s on s.uuid = a.status_uuid
where a.uuid = $1
  and c.uuid is not null
group by c.name, a.uuid, a.comment, c.point_amount, c.uuid, attachment_link, a.user_uuid, a.status_uuid,
         s.display_value;

-- name: GetAchievementSubCategories :many
select distinct c.uuid, c.name, cv.name as selected_value, cv.point, av_cv.name as available_value
from category c
         join achievement_category ac on ac.category_uuid = c.uuid
         join achievement_category_value acv on acv.achievement_uuid = ac.achievement_uuid
         join category_value cv on cv.uuid = acv.category_value_uuid and cv.category_uuid = c.uuid
         join category_value av_cv on av_cv.category_uuid = c.uuid
where parent_category is not null
  and ac.achievement_uuid = $1;

-- name: GetUserAchievements :many
select a.*,
       c.name                         as category_name,
       s.display_value                as status,
       c.uuid                         as category_uuid,
       coalesce(sum(cv.point) + c.point_amount, 0)::numeric as point_amount,
       count(*) over ()               as total_records
from achievement a
         join achievement_category ac on ac.achievement_uuid = a.uuid
         left join category c on c.uuid = ac.category_uuid and parent_category is null
         left join category c_p on c_p.uuid = ac.category_uuid and c_p.parent_category is not null
         join achievement_category_value acv on acv.achievement_uuid = a.uuid
         join category_value cv on cv.uuid = acv.category_value_uuid and cv.category_uuid = c.uuid
         join status s on s.uuid = a.status_uuid
where a.user_uuid = $1
  and s.internal_value != 'removed'
  and c.uuid is not null
group by c.name, a.uuid, a.comment, c.point_amount, c.uuid, attachment_link, a.user_uuid, a.status_uuid,
         s.display_value;

-- name: GetUserAchievementsWithPagination :many
select a.*,
       c.name                         as category_name,
       s.display_value                as status,
       c.uuid                         as category_uuid,
       coalesce(sum(cv.point) + c.point_amount, 0)::numeric as point_amount,
       count(*) over ()               as total_records
from achievement a
         join achievement_category ac on ac.achievement_uuid = a.uuid
         left join category c on c.uuid = ac.category_uuid and parent_category is null
         left join category c_p on c_p.uuid = ac.category_uuid and c_p.parent_category is not null
         join achievement_category_value acv on acv.achievement_uuid = a.uuid
         join category_value cv on cv.uuid = acv.category_value_uuid  and cv.category_uuid = c.uuid
         join status s on s.uuid = a.status_uuid
where a.user_uuid = $1
  and c.uuid is not null
group by c.name, a.uuid, a.comment, c.point_amount, c.uuid, attachment_link, a.user_uuid, a.status_uuid,
         s.display_value
limit $2 offset $3;

-- name: CreateAchievement :one
insert into achievement (comment, attachment_link, user_uuid, status_uuid)
values ($1, $2, $3,
        (select s.uuid from status s where s.internal_value = 'unapproved' and s.type = 'achievement_status'))
returning uuid;

-- name: CreateBatchAchievementCategory :batchexec
insert into achievement_category (category_uuid, achievement_uuid)
values ($1, $2);

-- name: UpdateAchievement :exec
update achievement
set comment         = $1,
    attachment_link = $2
where uuid = $3;

-- name: UpdateAchievementWithStatus :exec
update achievement a
set comment         = $1,
    attachment_link = $2,
    status_uuid = (select s.uuid from status s where s.internal_value = 'unapproved' and s.type = 'achievement_status')
where a.uuid = $3;

-- name: RemoveBatchAchievementCategory :batchexec
delete
from achievement_category
where category_uuid = $1
  and achievement_uuid = $2;

-- name: RemoveAllAchievementCategory :exec
delete
from achievement_category
where achievement_uuid = $1;

-- name: MakeAchievementUnapproved :exec
update achievement
set status_uuid = (select status.uuid from status where internal_value = 'unapproved' and type = 'achievement_status')
where achievement.uuid = $1;

-- name: MakeAchievementApproved :exec
update achievement
set status_uuid = (select status.uuid from status where internal_value = 'approved' and type = 'achievement_status')
where achievement.uuid = $1;

-- name: MakeAchievementUsed :exec
update achievement
set status_uuid = (select status.uuid from status where internal_value = 'used' and type = 'achievement_status')
where achievement.uuid = $1;

-- name: MakeAchievementDeclined :exec
update achievement
set status_uuid = (select status.uuid from status where internal_value = 'declined' and type = 'achievement_status')
where achievement.uuid = $1;

-- name: MakeAchievementRemoved :exec
update achievement
set status_uuid = (select status.uuid from status where internal_value = 'removed' and type = 'achievement_status')
where achievement.uuid = $1;

-- name: CreateAchievementCategoryValue :batchexec
insert into achievement_category_value (achievement_uuid, category_value_uuid)
VALUES ($1, (select cv.uuid from category_value cv where cv.category_uuid = $2 and cv.name = $3));

-- name: DeleteAchievementCategoryValueByAchievementUUID :exec
delete
from achievement_category_value
where achievement_uuid = $1;

