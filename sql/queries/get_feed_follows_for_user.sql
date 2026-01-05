-- name: GetFeedFollowsForUser :many
select f.name, u.name from feed_follows ff
    join feeds f on f.id = ff.feed_id
    join users u on u.id = ff.user_id
    where u.name  = $1;