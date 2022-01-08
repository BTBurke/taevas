package fs

import (
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed create.sql
var createSQL string

func New(root string, opts ...FSOption) (*Filesystem, error) {
	o := &option{
		file: ":memory:",
	}
	for _, opt := range opts {
		opt(o)
	}
	c, err := sqlx.Open("sqlite", o.file+"?_fk=1&_timeout=3000")
	if err != nil {
		return nil, err
	}
	if _, err := c.Exec(createSQL); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return &Filesystem{
		root: root,
		db:   c,
	}, nil
}

type option struct {
	file string
}

type FSOption func(o *option) error

func WithDBFile(path string) FSOption {
	return func(o *option) error {
		o.file = path
		return nil
	}
}
