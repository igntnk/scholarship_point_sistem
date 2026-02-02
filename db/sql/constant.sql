-- name: GetGradesConstant :one
select value
from constants
where name = 'available_student_grades';

-- name: ChangeGradesConstant :exec
update constants
set value = $1
where name = 'available_student_grades';