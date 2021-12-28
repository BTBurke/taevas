package db

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/BTBurke/taevas/utils"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Database struct {
	Conn *sqlx.DB
	FS   *Filesystem

	mu sync.Mutex
}

//go:embed create.sql
var createSQL string

func New() (*Database, error) {
	c, err := sqlx.Open("sqlite3", ":memory:?_fk=1&_timeout=3000")
	if err != nil {
		return nil, err
	}
	if _, err := c.Exec(createSQL); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return &Database{Conn: c}, nil
}

type Filesystem struct {
	root string
	db   *sqlx.DB
	mu   *sync.Mutex
}

func (fs *Filesystem) AddFile(name string) (int, error) {
	p := utils.ParsePath(name)
	d, ok := p.Depth()
	if !ok {
		return -1, fmt.Errorf("file exists outside go module: %s", name)
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var id int
	if err := fs.db.Get(&id, "INSERT INTO fs (dir, filename, depth) VALUES (?,?,?) RETURNING row_id", p.Dir(), p.FileName(), d); err != nil {
		return -1, err
	}
	return id, nil
}

func (fs *Filesystem) ReadDir(name string) ([]FSEntry, error) {
	p := utils.ParsePath(name)

	var entries []FSEntry
	if err := fs.db.Select(entries, ""); err != nil {
		return nil, err
	}
	return entries, nil
}

func (fs *Filesystem) ReadFile(name string) ([]byte, error) {
	var e FSEntry
	if err := fs.db.Get(&e, "SELECT * FROM filesystem WHERE path = ?", name); err != nil {
		return nil, fmt.Errorf("file %s not found: %w", name, err)
	}
	if e.Backing == 0 {
		return ioutil.ReadFile(filepath.Join(fs.root, name))
	}
	return e.data, nil
}

func (fs *Filesystem) Open(name string) (fs.Entry, error) {

}
