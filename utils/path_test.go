package utils

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	root := "/root/test"

	tt := []struct {
		name  string
		in    string
		dir   string
		fname string
		rel   string
		ext   string
		depth int
	}{
		{name: "root", in: root, dir: root, fname: "", rel: "", ext: "", depth: 0},
		{name: "root file", in: "test.tmpl", dir: root, fname: "test.tmpl", rel: "test.tmpl", ext: ".tmpl", depth: 0},
		{name: "subdir file", in: "test/test.tmpl", dir: root + "/test", fname: "test.tmpl", rel: "test/test.tmpl", ext: ".tmpl", depth: 1},
		{name: "subdir x2 file", in: "test/test2/test.tmpl", dir: root + "/test/test2", fname: "test.tmpl", rel: "test/test2/test.tmpl", ext: ".tmpl", depth: 2},
		{name: "subdir x3 file", in: "test/test2/test3/test.tmpl", dir: root + "/test/test2/test3", fname: "test.tmpl", rel: "test/test2/test3/test.tmpl", ext: ".tmpl", depth: 3},
		{name: "subdir x3 no file", in: "test/test2/test3", dir: root + "/test/test2/test3", fname: "", rel: "test/test2/test3", ext: "", depth: 3},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.in != root {
				tc.in = filepath.Join(root, tc.in)
			}

			p := ParsePath(tc.in)
			// set root manually for reproducability
			p.root = root

			assert.Equal(t, tc.dir, p.Dir())
			assert.Equal(t, tc.fname, p.FileName())
			rel, ok := p.RootRelative()
			assert.True(t, ok)
			assert.Equal(t, tc.rel, rel)
			d, ok := p.Depth()
			assert.True(t, ok)
			assert.Equal(t, tc.depth, d)
			assert.Equal(t, tc.ext, p.Ext())
			assert.Equal(t, strings.TrimRight(tc.fname, tc.ext), p.FileNameNoExt())
		})
	}
}
