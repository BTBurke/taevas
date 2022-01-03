package fs

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"sync"

	"github.com/BTBurke/taevas/utils"
	"github.com/jmoiron/sqlx"
)

type Filesystem struct {
	root string
	db   *sqlx.DB
	mu   sync.Mutex
}

// Add indexes a disk-backed file into the in-memory filesystem.  Subsequent operations on disk-backed files
// pass through to the underlying os.File implementation.
func (f *Filesystem) Add(name string) (int, error) {
	return f.add(name, 0, nil)
}

// AddVirtual creates a virtual in-memory entry for a virtual file.  Flush must be called to persist this virtual file
// to disk.  It may be operated on by while in memory.
func (f *Filesystem) AddVirtual(name string, data []byte) (int, error) {
	return f.add(name, 1, data)
}

func (f *Filesystem) add(name string, backing int, data []byte) (int, error) {
	p := utils.ParsePath(name)
	d, ok := p.Depth()
	if !ok {
		return -1, fmt.Errorf("file exists outside go module: %s", name)
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	var id int
	if err := f.db.Get(&id, "INSERT INTO fs (dir, filename, depth, data, backing) VALUES (?,?,?,?,?) RETURNING id", p.Dir(), p.FileName(), d, data, backing); err != nil {
		return -1, err
	}
	return id, nil
}

// Open returns an fs.File or a fs.PathError if the file does not exist
func (f *Filesystem) Open(name string) (fs.File, error) {
	var e Entry
	if err := f.db.Get(&e, "SELECT * FROM filesystem WHERE path = ? LIMIT 1", name); err != nil {
		return &Entry{}, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fmt.Errorf("lookup in fs database failed: %w", err),
		}
	}
	return &e, nil
}

// ReadDir returns all directory entries or a PathError
func (f *Filesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	p := utils.ParsePath(name)
	if !p.IsDir() || p.IsAbs() {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: name,
			Err:  fmt.Errorf("not a directory"),
		}
	}
	var e []Entry
	if err := f.db.Select(&e, "SELECT id, path, data, backing, time FROM read_dir WHERE dir = ?", name); err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: name,
			Err:  fmt.Errorf("error getting path listing from db: %w", err),
		}
	}
	out := make([]fs.DirEntry, len(e))
	for i, entry := range e {
		out[i] = entry
	}

	return out, nil
}

// ReadFile returns the file contents or a PathError
func (f *Filesystem) ReadFile(name string) ([]byte, error) {
	var e Entry
	if err := f.db.Get(&e, "SELECT * FROM filesystem WHERE path = ? LIMIT 1", name); err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: name,
			Err:  fmt.Errorf("error reading file from db: %w", err),
		}
	}
	b, err := ioutil.ReadAll(&e)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: name,
			Err:  fmt.Errorf("error reading bytes from file: %w", err),
		}
	}
	return b, nil
}

// Filesystem implements:
// fs.FS
// fs.ReadDirFS
// fs.ReadFileFS
var _ fs.FS = &Filesystem{}
var _ fs.ReadDirFS = &Filesystem{}
var _ fs.ReadFileFS = &Filesystem{}
