-- name: CreateRole :one
insert into auth_role (name)
values ($1)
returning uuid;

-- name: AddUsersToRole :batchexec
insert into user_role (role_uuid, user_uuid)
values ($1, $2);

-- name: CreateGroup :one
insert into auth_group (name)
values ($1)
returning uuid;

-- name: AddRolesToGroup :batchexec
insert into role_group (role_uuid, group_uuid)
values ($1, $2);

-- name: AddResourcesToGroup :batchexec
insert into group_resource (group_uuid, resource_uuid)
values ($1, $2);

-- name: UpdateRole :exec
update auth_role
set name = $1
where uuid = $2;

-- name: RemoveMembersFromRole :exec
DELETE
FROM user_role u
    USING UNNEST($1::uuid[]) AS ids(id)
WHERE u.uuid = ids.id;

-- name: UpdateGroup :exec
update auth_group
set name = $1
where uuid = $2;

-- name: RemoveRolesFromGroup :exec
DELETE
FROM role_group u
    USING UNNEST($1::uuid[]) AS ids(id)
WHERE u.uuid = ids.id;

-- name: RemoveResourcesFromGroup :exec
DELETE
FROM group_resource r
    USING UNNEST($1::uuid[]) AS ids(id)
WHERE r.uuid = ids.id;

-- name: DeleteMembersFromDeletedRole :exec
delete
from user_role
where role_uuid = $1;

-- name: DeleteRolesFromDeletedGroup :exec
delete
from role_group
where group_uuid = $1;

-- name: DeleteResourcesFromDeletedGroup :exec
delete
from group_resource
where group_uuid = $1;

-- name: DeleteRole :exec
delete
from auth_role
where uuid = $1;

-- name: DeleteGroup :exec
delete
from auth_group
where uuid = $1;

-- name: GetRoleByUUID :one
select *
from auth_role
where uuid = $1;

-- name: GetRoleMembers :many
select u.*
from sys_user u
         join user_role r on r.user_uuid = u.uuid
where role_uuid = $1;

-- name: GetGroupByUUID :one
select *
from auth_group
where uuid = $1;

-- name: GetGroupRoles :many
select r.*
from auth_role r
         join role_group g on r.uuid = g.role_uuid
where g.group_uuid = $1;

-- name: GetGroupResources :many
select r.*
from group_resource g
         join resource r on g.resource_uuid = r.uuid
where g.group_uuid = $1;

-- name: GetRoleList :many
select *
from auth_role;

-- name: GetGroupList :many
select *
from auth_group;

-- name: GetRoleListWithPagination :many
select *, count(*) over() total_records
from auth_role
limit $1 offset $2;

-- name: GetGroupListWithPagination :many
select *, count(*) over() total_records
from auth_group
limit $1 offset $2;