package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BTBurke/taevas/build/fs"
	"github.com/BTBurke/taevas/utils"
	"golang.org/x/net/html"
)

// Node is a single tag in an HTML document
type Node struct {
	name    string
	dir     string
	prev    *Node
	current *html.Node
}

// TemplateName is the name of the template currently being parsed
func (n *Node) TemplateName() string {
	return n.name
}

// TemplateDir is the directory in which the current template resides. This is
// useful for resolving relative paths.
func (n *Node) TemplateDir() string {
	return n.dir
}

// Attrs returns all the attributes of the current node
func (n *Node) Attrs() []html.Attribute {
	return n.current.Attr
}

// GetAttr returns the attribute referenced by key
func (n *Node) GetAttr(key string) (string, bool) {
	for _, attr := range n.current.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

// ReplaceAttr replaces the value at key with something else
func (n *Node) ReplaceAttr(key string, val string) {
	out := make([]html.Attribute, len(n.current.Attr))
	for i, attr := range n.current.Attr {
		switch {
		case attr.Key == key:
			out[i] = html.Attribute{
				Namespace: attr.Namespace,
				Key:       attr.Key,
				Val:       val,
			}
		default:
			out[i] = attr
		}
	}
	n.current.Attr = out
}

// RemoveAttr deletes the attribute at key
func (n *Node) RemoveAttr(key string) {
	for i, attr := range n.current.Attr {
		if attr.Key == key {
			n.current.Attr = append(n.current.Attr[0:i], n.current.Attr[i+1:]...)
			return
		}
	}
}

// AddAttr adds an attribute at key.  It does not verify that the attribute already
// exists.  Adding an attribute that already exists will result in multiple attributes of the
// same name.
func (n *Node) AddAttr(key, val string) {
	n.current.Attr = append(n.current.Attr, html.Attribute{
		Key: key,
		Val: val,
	})
}

// TagHandler performs some alteration of the tag referenced either by the Tag() value
// or the TagText.  Tag takes precendence over the TagText value.
type TagHandler interface {
	Tag()
	TagText() string
	Handle(n *Node) error
}

// TemplateCompiler compiles templates using the registered handlers
type TemplateCompiler interface {
	Scan() error
	RegisterTagHandler(h TagHandler) error
	Compile() error
}

// Context specifies the build context.  See docs for an explanation:
// TODO: add docs link
type Context struct {
	//
	TC       TemplateCompiler
	OutputFS map[string]*fs.Filesystem
	InputFS  *fs.Filesystem
}

// New returns a new build context, setting the template compiler and any global
// options
func New(root string, opts ...BuildOption) (*Context, error) {
	// TODO: figure out how to create a top level build context
	return nil, nil
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
