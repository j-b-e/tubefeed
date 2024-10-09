-- name: SaveMetadata :exec
INSERT INTO videos (
  uuid, title, channel, status, length, url, tabid
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
);

-- name: LoadDatabase :many
SELECT uuid, title, channel, status, length, url
FROM videos
WHERE tabid = ?;

-- name: GetVideo :one
SELECT title, channel, status, length, url
FROM videos
WHERE uuid = ?
LIMIT 1;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE uuid = ?;

-- name: DeleteVideosFromTab :exec
DELETE FROM videos
WHERE tabid = ?;

-- name: CountDuplicate :one
SELECT count(*)
FROM videos
WHERE url = ? AND tabid = ?;

-- name: LoadTabs :many
SELECT *
FROM tabs;

-- name: ChangeTabName :exec
UPDATE tabs
SET name = ?
WHERE id = ?;

-- name: AddTab :exec
INSERT INTO tabs (
  id, name
) VALUES (
  ?, ?
);

-- name: DeleteTab :exec
DELETE FROM tabs
WHERE id = ?;

-- name: GetLastTabId :one
SELECT id
FROM tabs
ORDER BY id DESC
LIMIT 1;
