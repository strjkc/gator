-- name: GetFeed :one
select * from feeds where url = $1;