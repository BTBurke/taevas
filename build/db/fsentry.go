package db

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/BTBurke/taevas/utils"
)

type FSEntry struct {
	ID      int `sql:"id"`
	Path    string
	Data    []byte
	Backing int
	Mod     time.Time `sql:"time"`

	buf  *bytes.Reader
	root string
}

func (e FSEntry) Name() string {
	p := utils.ParsePath(e.Path)
	if p.IsDir() {
		r := p.RootRelative()
		// returns the name of the last subdir or "."
		if r == "." || len(r) == 0 {
			return r
		}
		_, subdir := filepath.Split(r)
		return subdir
	}
	return p.FileName()
}

func (e FSEntry) IsDir() bool {
	p := utils.ParsePath(e.Path)
	return p.IsDir()
}

func (e FSEntry) Type() fs.FileMode {
	p := utils.ParsePath(e.Path)
	if p.IsDir() {
		return fs.ModeDir
	}
	return fs.ModePerm
}

func (e FSEntry) Size() int64 {
	return int64(len(e.Data))
}

func (e FSEntry) Mode() fs.FileMode {
	return e.Type()
}

func (e FSEntry) ModTime() time.Time {
	return e.Mod
}

func (e FSEntry) Sys() interface{} { return nil }

func (e FSEntry) Info() (fs.FileInfo, error) {
	if e.Backing == 0 {
		return os.Stat(filepath.Join(e.root, e.Path))
	}
	return e, nil
}

func (e FSEntry) Stat() (fs.FileInfo, error) {
	return e.Info()
}

func (e *FSEntry) Read(b []byte) (int, error) {
	if e.buf == nil {
		e.buf = bytes.NewReader(e.Data)
	}
	return e.buf.Read(b)
}

// FSEntry is a FileInfo and a DirEntry
var _ fs.FileInfo = FSEntry{}
var _ fs.DirEntry = FSEntry{}
