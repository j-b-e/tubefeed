 CREATE TABLE IF NOT EXISTS audio (
  id              uuid PRIMARY KEY,
  title           TEXT NOT NULL,
  channel         TEXT NOT NULL,
  length          INTEGER,  -- length is in seconds
  size            INTEGER,  -- size is in bytes
  source_url      TEXT NOT NULL,
  status          TEXT NOT NULL,
  provider_id     uuid,
  playlist_id     uuid,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP,
  FOREIGN KEY(playlist_id) REFERENCES playlist(uuid)
);

CREATE TABLE IF NOT EXISTS playlist (
  id              uuid PRIMARY KEY,
  name            TEXT NOT NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP
);
