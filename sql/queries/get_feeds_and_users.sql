-- name: GetFeedsAndUsers :many
SELECT f.name, f.url, u.name FROM feeds f
JOIN users u on f.user_id = u.id;