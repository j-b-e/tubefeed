-- name: SaveMetadata :exec
INSERT OR REPLACE INTO audio (
  uuid, title, channel, status, length, url, playlist_id
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
);

-- name: LoadDatabase :many
SELECT uuid, title, channel, status, length, url
FROM audio;

-- name: LoadPlaylist :many
SELECT uuid, title, channel, status, length, url
FROM audio
WHERE playlist_id = ?;

-- name: GetVideo :one
SELECT title, channel, status, length, url
FROM audio
WHERE uuid = ?
LIMIT 1;

-- name: DeleteVideo :exec
DELETE FROM audio
WHERE uuid = ?;

-- name: DeleteVideosFromTab :exec
DELETE FROM audio
WHERE playlist_id = ?;

-- name: CountDuplicate :one
SELECT count(*)
FROM audio
WHERE url = ? AND playlist_id = ?;

-- name: SetStatus :exec
UPDATE audio
SET status = ?
WHERE uuid = ?;

-- name: GetStatus :exec
SELECT status
FROM audio
WHERE uuid = ?;
