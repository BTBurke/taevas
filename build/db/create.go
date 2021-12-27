package db

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Database struct {
	Conn *sqlx.DB

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
