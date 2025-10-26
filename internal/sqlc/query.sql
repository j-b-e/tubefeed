-- name: SaveMetadata :exec
INSERT OR REPLACE INTO audio (
  id, title, channel, length, size, source_url, status, provider_id, playlist_id, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: AddPlaylist :exec
INSERT OR REPLACE INTO playlist (
  id, name
) VALUES (
  ?, ?
);

-- name: LoadDatabase :many
SELECT id, title, channel, status, length, source_url, playlist_id
FROM audio;

-- name: LoadAudioFromPlaylist :many
SELECT id, title, channel, status, length, source_url
FROM audio
WHERE playlist_id = ?;

-- name: LoadPlaylist :one
SELECT id, name
FROM playlist
WHERE id = ?;

-- name: GetAudio :one
SELECT title, channel, status, playlist_id, playlist.name as playlist_name, length, source_url
FROM audio
JOIN playlist ON audio.playlist_id = playlist.id
WHERE audio.id = ?
LIMIT 1;

-- name: DeleteAudio :exec
DELETE FROM audio
WHERE id = ?;

-- name: DeleteAudioFromPlaylist :exec
DELETE FROM audio
WHERE playlist_id = ?;

-- name: CountDuplicate :one
SELECT count(*)
FROM audio
WHERE source_url = ? AND playlist_id = ?;

-- name: SetStatus :exec
UPDATE audio
SET status = ?
WHERE id = ?;

-- name: GetStatus :exec
SELECT status
FROM audio
WHERE id = ?;
