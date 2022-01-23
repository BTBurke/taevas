package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BTBurke/taevas/build/fs"
	"github.com/BTBurke/taevas/utils"
)

type Context struct {
	// a lookup table where the key is the target to render and a list of possible template references that are resolvable from the target
	Targets map[string][]string

	options options
	in      *fs.Filesystem
	out     *fs.Filesystem
}

func New(root string, opts ...BuildOption) (*Context, error) {
	// sets up defaults which can be overridden by later options
	allOpts := append([]BuildOption{
		withRoot(root),
		WithOutputDirectory("dist", true),
		WithTemplateExtension(".tmpl"),
		WithTimeout(30 * time.Second),
	}, opts...)

	var o options
	for _, opt := range allOpts {
		if err := opt(&o); err != nil {
			return nil, fmt.Errorf("error setting build option: %w", err)
		}
	}

}

type BuildOption func(*options) error

type options struct {
	templateExt     string
	root            string
	outDir          string
	outDirOverwrite bool
	timeout         time.Duration
}

func WithTemplateExtension(ext string) BuildOption {
	return func(o *options) error {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		o.templateExt = ext
		return nil
	}
}

func withRoot(root string) BuildOption {
	return func(o *options) error {
		// root can be relative or absolute directory. If relative, it should refer to Go module root
		p := utils.ParsePath(root)
		if !p.IsAbs() {
			root = filepath.Join(utils.GoRoot(), root)
		}

		if _, err := os.Stat(root); os.IsNotExist(err) {
			return fmt.Errorf("error setting build options: root %s does not exist (%w)", root, err)
		}
		o.root = root
		return nil
	}
}

func WithOutputDirectory(outputDirectory string, overwrite bool) BuildOption {
	return func(o *options) error {
		p := utils.ParsePath(outputDirectory)
		if !p.IsAbs() {
			outputDirectory = filepath.Join(utils.GoRoot(), outputDirectory)
		}
		o.outDir = outputDirectory
		o.outDirOverwrite = overwrite
		return nil
	}
}

func WithTimeout(d time.Duration) BuildOption {
	return func(o *options) error {
		o.timeout = d
		return nil
	}
}
