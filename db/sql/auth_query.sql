-- name: ChangePassword :exec
update sys_user set password = $1, salt = $2 where uuid = $3;

-- name: CheckHasPermission :one
select u.uuid
from sys_user u
         join user_role ur on u.uuid = ur.user_uuid
         join role_group rg on rg.role_uuid = ur.role_uuid
         join group_resource gr on gr.group_uuid = rg.group_uuid
         join resource r on gr.resource_uuid = r.uuid
where u.uuid = $1
  and r.value = $2;