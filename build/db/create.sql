CREATE TABLE IF NOT EXISTS config (
  key TEXT NOT NULL,
  value BLOB NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS cfg_idx ON config(key);

INSERT INTO config (key, value) VALUES ('template_extension', '.tmpl');


-- creates filesystem view, optionally storing the content of virtual files in data
-- backing indicates whether the file exists on disk or purely in memory: disk backed = 0; virtual = 1
-- depth indicates the subdirectory depth relative to the module root (root = 0)
-- entries are not created for directories, they implicitly exist when a file is created
CREATE TABLE IF NOT EXISTS fs (
  dir TEXT NOT NULL CHECK(length(dir) > 0),
  filename TEXT NOT NULL CHECK(length(filename) > 0),
  data BLOB,
  depth INTEGER NOT NULL DEFAULT 0 CHECK(depth >= 0),
  backing INTEGER NOT NULL DEFAULT 0 CHECK(backing >= 0 AND backing <= 1),
  modtime INTEGER NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS full_path_idx ON fs(dir,filename);

CREATE VIEW IF NOT EXISTS filesystem AS 
  SELECT 
    row_id as id, 
    dir || '/' || filename as path,
    data,
    backing,
    (SELECT datetime(modtime, 'unixepoch', 'localtime')) as time
    (SELECT value FROM config WHERE key = 'root') || '/' || dir || '/' || filename as abs_path
  FROM
    fs;

CREATE VIEW IF NOT EXISTS directories AS
  SELECT DISTINCT dir, depth FROM fs;
 
-- templates stored in this table reference an fs entry. Types range from -1 to 3: 
-- unknown = -1; layout = 0; global partial = 1; local partial = 2; target = 3
CREATE TABLE IF NOT EXISTS templates (
  ttype INTEGER NOT NULL DEFAULT -1 CHECK(ttype >= -1 AND ttype <= 3),
  fsentry INTEGER NOT NULL,
  FOREIGN KEY(fsentry) REFERENCES fs(row_id)
);

CREATE VIEW IF NOT EXISTS layouts AS 
  SELECT * FROM templates INNER JOIN fs ON templates.fsentry = fs.row_id WHERE templates.ttype = 0;

CREATE VIEW IF NOT EXISTS globals AS
  SELECT * FROM templates INNER JOIN fs ON templates.fsentry = fs.row_id WHERE templates.ttype = 1;

CREATE VIEW IF NOT EXISTS locals AS
 SELECT * FROM templates INNER JOIN fs ON templates.fsentry = fs.row_id WHERE templates.ttype = 2;

CREATE VIEW IF NOT EXISTS targets AS
  SELECT * FROM templates INNER JOIN fs ON templates.fsentry = fs.row_id WHERE templates.ttype = 3;
