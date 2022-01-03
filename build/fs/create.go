package fs

import (
	_ "embed"
	"fmt"

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
