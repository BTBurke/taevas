package fs

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/BTBurke/taevas/utils"
)

type Entry struct {
	ID      int `sql:"id"`
	Path    string
	Data    []byte
	Backing int
	Time    int64 `sql:"time"`

	buf  *bytes.Reader
	fh   *os.File
	root string
}

func (e Entry) Stat() (fs.FileInfo, error) {
	if e.Backing == 0 {
		return os.Stat(filepath.Join(e.root, e.Path))
	}
	return e, nil
}

func (e Entry) Name() string {
	p := utils.ParsePath(e.Path)
	if p.IsDir() {
		r := p.Dir()
		// returns the name of the last subdir or "."
		if r == "." || len(r) == 0 {
			return "."
		}
		_, subdir := filepath.Split(r)
		return subdir
	}
	return p.FileName()
}

func (e Entry) IsDir() bool {
	p := utils.ParsePath(e.Path)
	return p.IsDir()
}

func (e Entry) Type() fs.FileMode {
	p := utils.ParsePath(e.Path)
	if p.IsDir() {
		return 0755
	}
	return 0644
}

func (e Entry) Size() int64 {
	return int64(len(e.Data))
}

func (e Entry) Mode() fs.FileMode {
	return e.Type()
}

func (e Entry) ModTime() time.Time {
	return time.Unix(e.Time, 0)
}

func (e Entry) Sys() interface{} { return nil }

func (e Entry) Info() (fs.FileInfo, error) {
	if e.Backing == 0 {
		return os.Stat(filepath.Join(e.root, e.Path))
	}
	return e.Stat()
}

func (e *Entry) Read(b []byte) (int, error) {
	if e.Backing == 0 {
		if e.fh == nil {
			fh, err := os.Open(filepath.Join(e.root, e.Path))
			if err != nil {
				return 0, err
			}
			e.fh = fh
			return e.fh.Read(b)
		}
		return e.fh.Read(b)
	}
	if e.buf == nil {
		e.buf = bytes.NewReader(e.Data)
	}
	return e.buf.Read(b)
}

func (e *Entry) Close() error {
	if e.Backing == 0 && e.fh != nil {
		return e.fh.Close()
	}
	if e.buf != nil {
		e.buf = nil
		return nil
	}
	return nil
}

// Entry implements:
// fs.FileInfo
// fs.DirEntry
// fs.File
var _ fs.FileInfo = Entry{}
var _ fs.DirEntry = Entry{}
var _ fs.File = &Entry{}
