-- name: getProfile :many
SELECT users.user_id,
    users.quote,
    users.date as roll_date,
    users.favorite,
    users.tokens,
    users.anilist_url,
    characters.id,
    characters.image,
    characters.name,
    characters.date,
    characters.type
FROM users
    INNER JOIN characters ON characters.user_id = users.user_id
WHERE users.user_id = $1;
-- name: getUserByAnilist :one
SELECT users.user_id,
    users.quote,
    users.date as roll_date,
    users.favorite,
    users.tokens,
    users.anilist_url
FROM users
WHERE users.anilist_url = $1;