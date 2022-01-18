package build

import (
	"context"
	"strings"

	"github.com/BTBurke/taevas/build/fs"
)

type Context struct {
	// a lookup table where the key is the target to render and a list of possible template references that are resolvable from the target
	Targets map[string][]string

	options options
	in      *fs.Filesystem
	out     *fs.Filesystem
	ctx     context.Context
}

type BuildOption func(*options) error

type options struct {
	templateExt string
	root        string
	outDir      string
}

var DefaultOptions = options{
	templateExt: ".tmpl",
	outDir:      "dist",
}

func WithTemplateExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

}
