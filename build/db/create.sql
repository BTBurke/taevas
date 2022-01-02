-- keeps configuration and default values
CREATE TABLE IF NOT EXISTS config (
  key TEXT NOT NULL,
  value BLOB NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS cfg_idx ON config(key);

INSERT INTO config (key, value) VALUES ('template_extension', '.tmpl');

-- convenience view to return the template extension
CREATE VIEW IF NOT EXISTS template_extension AS
  SELECT value FROM config WHERE key = 'template_extension' LIMIT 1;

-- creates filesystem view, optionally storing the content of virtual files in data
-- backing indicates whether the file exists on disk or purely in memory: disk backed = 0; virtual = 1
-- depth indicates the subdirectory depth relative to the module root (root = 0)
-- entries are not created for directories, they implicitly exist when a file is created
CREATE TABLE IF NOT EXISTS fs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  dir TEXT NOT NULL CHECK(length(dir) > 0),
  filename TEXT NOT NULL CHECK(length(filename) > 0),
  data BLOB,
  depth INTEGER NOT NULL DEFAULT 0 CHECK(depth >= 0),
  backing INTEGER NOT NULL DEFAULT 0 CHECK(backing >= 0 AND backing <= 1),
  modtime INTEGER NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS full_path_idx ON fs(dir,filename);

-- simplified view of the filesystem
CREATE VIEW IF NOT EXISTS filesystem AS 
  SELECT 
    id, 
    dir || '/' || filename as path,
    data,
    backing,
    (SELECT datetime(modtime, 'unixepoch', 'localtime')) as time
  FROM
    fs;

-- view of all unique directories with real or virtual files
CREATE VIEW IF NOT EXISTS directories AS
  SELECT DISTINCT dir, depth FROM fs;

-- finds layout templates that look start with a _.  Layouts may also inherit from other layouts using
-- _layout.parent.tmpl
CREATE VIEW IF NOT EXISTS layouts AS 
  SELECT * FROM fs WHERE filename LIKE '__%' || (SELECT * FROM template_extension) ESCAPE '_';

-- returns the short name of the template _layout.tmpl -> layout
CREATE VIEW IF NOT EXISTS layouts_short_name AS
  SELECT
    id,
    filename,
    SUBSTR(filename, 2, INSTR(filename, '.')-2) as short_name
  FROM layouts;

-- Target templates are of the form <name>.<layout>.tmpl.  Targets are the only template that is directly
-- rendered.  Convenience functions are created to render these targets from http.Handlers.
CREATE VIEW IF NOT EXISTS targets AS
  SELECT * FROM fs WHERE filename LIKE '%.%' || (SELECT * FROM template_extension)
  AND id NOT IN (SELECT id FROM layouts);

-- Locals are partial templates in the same directory as a target template.  Locals are placed in the parse tree
-- with their associated targets in the same package.  This allows per-package partial includes that do not affect templates
-- in other directories/packages.
CREATE VIEW IF NOT EXISTS locals AS
  SELECT * FROM fs WHERE filename LIKE '%' || (SELECT * FROM template_extension) 
  AND id NOT IN (SELECT id FROM layouts)
  AND id NOT IN (SELECT id FROM targets)
  AND dir IN (SELECT dir FROM targets);

-- Globals are partial templates that are in directories with no targets.  These are placed in the parse tree for every
-- target.  This is useful for global templates such as UI components.
CREATE VIEW IF NOT EXISTS globals AS
  SELECT * FROM fs WHERE filename LIKE '%' || (SELECT * FROM template_extension) 
  AND id NOT IN (SELECT id FROM layouts)
  AND id NOT IN (SELECT id FROM targets)
  AND id NOT IN (SELECT id FROM locals);

-- Finds the layout associated with the target.  The first layout that matches walking from the target
-- back to the module root is the layout used.
CREATE VIEW IF NOT EXISTS target_layout AS
  SELECT 
    targets.id as target_id, 
    targets.dir || '/' || targets.filename as target_path,
    layouts.id as layout_id,
    layouts.dir || '/' || layouts.filename as layout_path    
    FROM targets JOIN layouts
    ON targets.filename LIKE '%.' || (SELECT short_name FROM layouts_short_name WHERE id = layouts.id) || '.%'
    WHERE targets.dir LIKE LTRIM(layouts.dir, '.') || '%'
    ORDER BY layouts.depth DESC;

-- Finds all local partial templates in the same directory as the target
CREATE VIEW IF NOT EXISTS target_locals AS
  SELECT
    targets.id as target_id,
    targets.dir || '/' || targets.filename as target_path,
    locals.id as local_id,
    locals.dir || '/' || locals.filename as local_path
    FROM targets JOIN locals
    ON targets.dir = locals.dir;

-- Finds all global templates that should be included in the parse tree.  This is a convenience view to easily generate a tree of 
-- templates to be parsed.
CREATE VIEW IF NOT EXISTS target_globals AS
  SELECT
    targets.id as target_id,
    targets.dir || '/' || targets.filename as target_path,
    globals.id as global_id,
    globals.dir || '/' || globals.filename as global_path
    FROM targets CROSS JOIN globals;

-- Finds the parent of a layout if it inherits from another layout.  This is used later in a recursive table expression to walk
-- an arbitrary tree of layout template inheritance.
CREATE VIEW IF NOT EXISTS layout_parent AS
  SELECT 
    l1.id as layout_id, 
    l1.dir || '/' || l1.filename as layout_path,
    l2.id as parent_id,
    l2.dir || '/' || l2.filename as parent_path    
    FROM layouts l1 JOIN layouts l2 
    ON l1.filename LIKE '%.' || (SELECT short_name FROM layouts_short_name WHERE id = l2.id) || '.%'
    WHERE l1.dir LIKE LTRIM(l2.dir, '.') || '%'
    ORDER BY l2.depth DESC;

-- A recursive view that yields a tree of layouts needed to render a target
CREATE VIEW IF NOT EXISTS layout_tree AS 
  WITH RECURSIVE layout_cte(target_path, layout_path, ord) AS (
    SELECT target_path, layout_path, (SELECT 0) as ord
    FROM target_layout 

    UNION ALL

    SELECT cte.target_path, lp.parent_path, cte.ord + 1 as ord
    FROM layout_cte cte JOIN layout_parent lp ON cte.layout_path = lp.layout_path
  )
  SELECT * FROM layout_cte ORDER BY ord DESC; 

-- Yields the complete tree of templates that need to be parsed to render a target.  The precedence of templates matters:
-- Layout (may be multiple if they inherit from each other) -> globals -> locals -> target
-- This is the only view really needed to render a particular target.  It makes use of all the other views to determine a per-target
-- tree of all required templates.
CREATE VIEW IF NOT EXISTS target_tree AS
  SELECT target_path, layout_path as template_path FROM layout_tree
  UNION ALL
  SELECT target_path, global_path as template_path FROM target_globals
  UNION ALL
  SELECT target_path, local_path as template_path FROM target_locals;

