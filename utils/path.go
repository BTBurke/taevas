package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Path represents an arbitrary location on disk, either a directory or file
type Path struct {
	dir  string
	file string
}

// ParsePath returns a path by parsing a string that may contain a file
func ParsePath(p string) Path {
	p = filepath.Clean(p)
	if p == "." {
		return Path{
			dir: p,
		}
	}
	d, f := filepath.Split(p)
	// if no extension exists, the whole thing was a directory
	if filepath.Ext(f) == "" {
		d = p
		f = ""
	}
	// if no directory exists, it's a file in the root
	if d == "" {
		d = "."
	}
	return Path{
		dir:  d,
		file: f,
	}
}

// New constructs a path from a directory and file
func NewPath(dir string, file string) Path {
	dir = filepath.Clean(dir)
	return Path{
		dir:  dir,
		file: file,
	}
}

// Dir returns the directory
func (p Path) Dir() string {
	return strings.TrimRight(p.dir, "/")
}

// FileName returns the file name with extension
func (p Path) FileName() string {
	return p.file
}

// Ext returns the file extension
func (p Path) Ext() string {
	return filepath.Ext(p.file)
}

// FileNameNoExt returns the file name without extension
func (p Path) FileNameNoExt() string {
	return strings.TrimRight(p.file, p.Ext())
}

// IsAbs checks if the path directory is absolute
func (p Path) IsAbs() bool {
	return filepath.IsAbs(p.dir)
}

// IsRoot checks if the current path is at the module root
func (p Path) IsRoot() bool {
	return p.dir == "."
}

// ParentDirectory walks up one directory level and returns a new path pointing to the directory
func (p Path) ParentDirectory() (Path, error) {
	up := filepath.Dir(p.dir)
	if up == "." || up == "/" {
		return Path{}, errors.New("failed to move up one directory level")
	}
	return NewPath(up, ""), nil
}

// IsDir tests whether the path is a directory
func (p Path) IsDir() bool {
	return p.dir != "" && p.file == ""
}

// Exists checks if the file exists at path
func (p Path) Exists() bool {
	if _, err := os.Stat(p.String()); !errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}

// RootRelative returns the relative path from the module root
func (p Path) RootRelative() string {
	if !p.IsAbs() {
		return strings.TrimRight(filepath.Join(p.dir, p.FileName()), "/")
	}
	r, ok := p.rootRelativeDir()
	if !ok {
		return strings.TrimRight(filepath.Join(p.dir, p.FileName()), "/")
	}
	return strings.TrimRight(filepath.Join(r, p.FileName()), "/")
}

// rootRelativeDir returns the root relative directory without filename
func (p Path) rootRelativeDir() (string, bool) {
	if !p.IsAbs() {
		return p.dir, true
	}
	r, err := filepath.Rel(GoRoot(), p.dir)
	if err != nil {
		return p.dir, false
	}
	return strings.TrimLeft(r, "."), true
}

// Depth returns the package level depth starting from the module root
func (p Path) Depth() (int, bool) {
	rr, ok := p.rootRelativeDir()
	if !ok {
		return 0, false
	}
	// special case for root directory
	if rr == "." {
		return 0, true
	}
	rr = strings.TrimRight(rr, string(filepath.Separator))
	return len(strings.Split(rr, string(filepath.Separator))), true
}

// String returns the absolute path
func (p Path) String() string {
	if p.dir == "." {
		return p.dir + "/" + p.file
	}
	return filepath.Join(p.dir, p.file)
}
