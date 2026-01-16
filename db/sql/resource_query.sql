-- name: BatchInsertResources :batchexec
insert into resource (value) values ($1);

-- name: BatchDeleteResources :exec
DELETE FROM resource u
    USING UNNEST($1::uuid[]) AS ids(id)
WHERE u.uuid = ids.id;

-- name: GetResourceList :many
select * from resource;