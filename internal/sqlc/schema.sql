 CREATE TABLE IF NOT EXISTS audio (
  uuid            TEXT PRIMARY KEY,
  title           TEXT NOT NULL,
  channel         TEXT NOT NULL,
  length          INTEGER,  -- length is in seconds
  size            INTEGER,  -- size is in bytes
  url             TEXT NOT NULL,
  status          TEXT NOT NULL,
  provider_id     TEXT,
  playlist_id     TEXT,
  created_at      TIMESTAMP DATETIME,
  FOREIGN KEY(playlist_id) REFERENCES playlist(uuid)
);

CREATE TABLE IF NOT EXISTS playlist (
  uuid            TEXT PRIMARY KEY,
  name            TEXT NOT NULL
);
