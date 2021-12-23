package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Path represents an arbitrary location on disk, either a directory or file
type Path struct {
	root string
	dir  string
	file string
}

// ParsePath returns a path by parsing a string that may contain a file
func ParsePath(p string) Path {
	p = filepath.Clean(p)
	d, f := filepath.Split(p)
	// if no extension exists, the whole thing was a directory
	if filepath.Ext(f) == "" {
		d = p
		f = ""
	}
	return Path{
		root: GoRoot(),
		dir:  d,
		file: f,
	}
}

// New constructs a path from a directory and file
func NewPath(dir string, file string) Path {
	dir = filepath.Clean(dir)
	return Path{
		root: GoRoot(),
		dir:  dir,
		file: file,
	}
}

// Dir returns the directory
func (p Path) Dir() string {
	return p.dir
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
	return p.root == p.dir
}

// ParentDirectory walks up one directory level and returns a new path pointing to the directory
func (p Path) ParentDirectory() (Path, error) {
	if p.IsRoot() {
		return Path{}, errors.New("at go root, cannot walk up one level")
	}
	if !p.IsAbs() && len(strings.Split(p.dir, string(filepath.Separator))) == 1 {
		return Path{}, errors.New("relative directory top leve, cannot walk up one level")
	}
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
func (p Path) RootRelative() (string, bool) {
	if !p.IsAbs() {
		return "", false
	}
	r, err := filepath.Rel(p.root, p.dir)
	if err != nil {
		return "", false
	}
	return r, true
}

// Depth returns the package level depth starting from the module root
func (p Path) Depth() (int, bool) {
	rr, ok := p.RootRelative()
	if !ok {
		return 0, false
	}
	return len(strings.Split(rr, string(filepath.Separator))), true
}

// String returns the absolute path
func (p Path) String() string {
	if p.IsAbs() {
		return filepath.Join(p.dir, p.file)	
	} else {
		return filepath.Join(p.root, p.dir, p.file)
	}	
}

