-- name: GetUser :one
Select * from users where name = $1;
