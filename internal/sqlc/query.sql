-- name: CreatePlaylist :exec
INSERT OR REPLACE INTO playlist (
  id, name
) VALUES (
  ?, ?
);

-- name: UpdatePlaylist :exec
UPDATE playlist
SET name = ?
WHERE id = ?;

-- name: DeletePlaylist :exec
DELETE FROM playlist
WHERE id = (?);

-- name: ListPlaylist :many
SELECT *
FROM playlist
ORDER BY id;

-- name: GetPlaylist :one
SELECT *
FROM playlist
WHERE id = ?;

-- name: CountPlaylist :one
SELECT count(*)
FROM playlist;

-- name: CreateAudio :exec
INSERT OR REPLACE INTO audio (
  id, title, channel, length, size, source_url, status, provider_id, playlist_id, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: DeleteAudio :exec
DELETE FROM audio
WHERE id = ?;

-- name: UpdateAudio :exec
UPDATE audio
SET
   title = ?, channel = ?, length = ?,
   size = ?, source_url = ?, status = ?,
   provider_id = ?, playlist_id = ?, updated_at = ?
WHERE id = ?;

-- name: ListAudio :many
SELECT *
FROM audio;

-- name: LoadAudioByPlaylist :many
SELECT *
FROM audio
WHERE playlist_id = ?;

-- name: GetAudio :one
SELECT *, playlist.name as playlist_name
FROM audio
JOIN playlist ON audio.playlist_id = playlist.id
WHERE audio.id = ?
LIMIT 1;

-- name: CountAudio :one
SELECT count(*)
FROM audio;

-- name: CountDuplicate :one
SELECT count(*)
FROM audio
WHERE source_url = ? AND playlist_id = ?;
