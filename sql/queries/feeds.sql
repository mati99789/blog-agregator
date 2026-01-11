-- name: GetFeeds :many
SELECT *
FROM feeds;

-- name: GetFeedByID :one
SELECT *
FROM feeds
WHERE id = $1;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, NOW(), NOW(), $2, $3, $4)
RETURNING *;

-- name: UpdateFeed :one
UPDATE feeds
SET url        = $1,
    name       = $2,
    updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: DeleteFeed :exec
DELETE
FROM feeds
WHERE id = $1;


-- name: GetAllFeedsWithUser :many
SELECT feeds.name AS feed_name, feeds.url, users.name AS user_name
FROM feeds
         INNER JOIN users on feeds.user_id = users.id;

-- name: GetFeedByUrl :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: MarkFeedFetched :exec 
UPDATE
    feeds
SET last_fetched_at = now(),
    updated_at      = now()
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;