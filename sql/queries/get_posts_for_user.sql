-- name: GetPostsForUser :many
select * from posts where feed_id in (
SELECT feed_id from feed_follows where user_id = $1) order by published_at limit $2;
