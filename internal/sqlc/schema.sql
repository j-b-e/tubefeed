 CREATE TABLE IF NOT EXISTS videos (
  uuid            TEXT PRIMARY KEY,
  title           TEXT NOT NULL,
  channel         TEXT NOT NULL,
  status          TEXT NOT NULL,
  length          INTEGER NOT NULL,  -- length is in seconds
  size            INTEGER,  -- size is in bytes
  url             TEXT NOT NULL,
  tabid           INTEGER,
  FOREIGN KEY(tabid) REFERENCES tabs(id)
);

CREATE TABLE IF NOT EXISTS tabs (
  id    INTEGER PRIMARY KEY,
  name  TEXT NOT NULL
);
