-- name: GetSimpleUserAchievementByUUID :one
select a.*, c.name as category_name, s.display_value as status, sum(c_p.point_amount) as point_amount
from achievement a
         join user_achievement ua on a.uuid = ua.achievement_uuid
         join achievement_category ac on ac.achievement_uuid = a.uuid
         join category c on c.uuid = ac.category_uuid and parent_category is null
         join category c_p on c.uuid = ac.category_uuid
         join status s on s.uuid = a.status_uuid
where a.uuid = $1
group by c.name, a.uuid, a.comment, attachment_link, a.user_uuid, a.status_uuid, s.display_value;

-- name: GetUserAchievements :many
select a.*, c.name as category_name, s.display_value as status, sum(c_p.point_amount) as point_amount
from achievement a
         join user_achievement ua on a.uuid = ua.achievement_uuid
         join achievement_category ac on ac.achievement_uuid = a.uuid
         join category c on c.uuid = ac.category_uuid and parent_category is null
         join category c_p on c.uuid = ac.category_uuid
         join status s on s.uuid = a.status_uuid
where a.user_uuid = $1
group by c.name, a.uuid, a.comment, attachment_link, a.user_uuid, a.status_uuid, s.display_value;

-- name: GetUserAchievementsWithPagination :many
select a.*, c.name as category_name, s.display_value as status, sum(c_p.point_amount) as point_amount, count(*) over() as total_records
from achievement a
         join user_achievement ua on a.uuid = ua.achievement_uuid
         join achievement_category ac on ac.achievement_uuid = a.uuid
         join category c on c.uuid = ac.category_uuid and parent_category is null
         join category c_p on c.uuid = ac.category_uuid
         join status s on s.uuid = a.status_uuid
where a.user_uuid = $1
group by c.name, a.uuid, a.comment, attachment_link, a.user_uuid, a.status_uuid, s.display_value
    limit $2 offset $3;

-- name: CreateAchievement :one
insert into achievement (comment, attachment_link, user_uuid, status_uuid)
values ($1, $2, $3, (select s.uuid from status s where s.internal_value = 'unapproved' and s.type = 'achievement_status'))
    returning uuid;

-- name: CreateBatchAchievementCategory :batchexec
insert into achievement_category (category_uuid, achievement_uuid)
values ($1, $2);

-- name: UpdateAchievement :exec
update achievement
set comment         = $1,
    attachment_link = $2
where uuid = $3;

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

