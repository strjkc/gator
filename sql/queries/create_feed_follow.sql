-- name: CreateFeedFollow :many
with ins as(
    insert into feed_follows(created_at, updated_at, user_id, feed_id)
    values(
           $1,
           $2,
           $3,
           $4
    )
    RETURNING *
)
select * from ins ff join feeds f on f.id = ff.feed_id
           join users u on ff.user_id = u.id;
