-- name: MarkedFeedFetched :one
update feeds
    set updated_at = current_timestamp, last_fetched_at = current_timestamp
    where id = $1

returning *;