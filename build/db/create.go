package db

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/BTBurke/taevas/utils"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed create.sql
var createSQL string

func New(root string) (*Filesystem, error) {
	c, err := sqlx.Open("sqlite", ":memory:?_fk=1&_timeout=3000")
	if err != nil {
		return nil, err
	}
	if _, err := c.Exec(createSQL); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return &Filesystem{root: root, db: c}, nil
}

type Filesystem struct {
	root string
	db   *sqlx.DB
	mu   sync.Mutex
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
	if err := fs.db.Get(&id, "INSERT INTO fs (dir, filename, depth) VALUES (?,?,?) RETURNING id", p.Dir(), p.FileName(), d); err != nil {
		return -1, err
	}
	return id, nil
}
