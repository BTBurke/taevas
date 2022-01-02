package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddFiles(t *testing.T) {
	tt := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{name: "mod root", path: "_layout.tmpl", shouldErr: false},
		{name: "abs path", path: "/test.tmpl", shouldErr: true},
		{name: "local 1", path: "a/1.tmpl", shouldErr: false},
		{name: "local 2", path: "a/2.tmpl", shouldErr: false},
		{name: "global 1", path: "g/1.tmpl", shouldErr: false},
		{name: "global 2", path: "g/2.tmpl", shouldErr: false},
		{name: "target 1", path: "a/target1.layout.tmpl", shouldErr: false},
	}

	fs, err := New("/test")
	require.NoError(t, err)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			id, err := fs.AddFile(tc.path)
			switch tc.shouldErr {
			case true:
				assert.Error(t, err)
			default:
				assert.NoError(t, err)
				assert.LessOrEqual(t, 0, id)
			}
		})
	}
}

func TestDetectTemplates(t *testing.T) {
	tt := []struct {
		name string
		path string
	}{
		{name: "layout", path: "_layout.tmpl"},
		{name: "nested layout", path: "a/_sub.layout.tmpl"},
		{name: "local 1", path: "a/1.tmpl"},
		{name: "local 2", path: "a/2.tmpl"},
		{name: "global 1", path: "g/1.tmpl"},
		{name: "global 2", path: "g/2.tmpl"},
		{name: "target 1", path: "a/target1.layout.tmpl"},
		{name: "target 2", path: "a/target2.sub.tmpl"},
	}

	expect := map[string][]string{
		"layouts": {"./_layout.tmpl", "a/_sub.layout.tmpl"},
		"shorts":  {"layout", "sub"},
		"locals":  {"a/1.tmpl", "a/2.tmpl"},
		"globals": {"g/1.tmpl", "g/2.tmpl"},
		"targets": {"a/target1.layout.tmpl", "a/target2.sub.tmpl"},
	}

	fs, err := New("/test")
	require.NoError(t, err)

	// populate in memory file system
	for _, tc := range tt {
		if _, err := fs.AddFile(tc.path); err != nil {
			assert.NoError(t, err)
		}
	}

	// sql queries should auto categorize different types of templates
	for _, ttype := range []string{"layouts", "targets", "globals", "locals"} {
		t.Run(ttype, func(t *testing.T) {
			var templates []string
			assert.NoError(t, fs.db.Select(&templates, "SELECT dir || '/' || filename FROM "+ttype))
			assert.Equal(t, expect[ttype], templates)
			t.Logf("%s: %+v", ttype, templates)
		})
	}

	// helper tables record short names for the layout (removing initial _ and trailing extension plus any parent layout)
	for i, layout := range []string{"_layout.tmpl", "_sub.layout.tmpl"} {
		var short string
		assert.NoError(t, fs.db.Get(&short, "SELECT short_name FROM layouts_short_name WHERE filename = ? LIMIT 1", layout))
		assert.Equal(t, expect["shorts"][i], short)
		t.Logf("short - %s:%s", layout, short)
	}

	// a target should be able to find its possible parent layouts
	// NOTE: final determination should be made on the basis of path and depth
	for i, target := range expect["targets"] {
		var parents []string
		assert.NoError(t, fs.db.Select(&parents, "SELECT layout_path FROM target_layout WHERE target_path = ?", target))
		assert.Equal(t, expect["layouts"][i], parents[0])
		t.Logf("%s: %v", target, parents)
	}

	// a layout can have a parent layout
	for i, lp := range []string{"a/_sub.layout.tmpl"} {
		var parents []string
		assert.NoError(t, fs.db.Select(&parents, "SELECT parent_path FROM layout_parent WHERE layout_path = ?", lp))
		assert.Equal(t, expect["layouts"][i], parents[0])
		t.Logf("%s: %v", lp, parents)
	}

	// a nested layout should yield the full tree of layout templates required for the target
	expectTree := []string{"./_layout.tmpl", "a/_sub.layout.tmpl"}
	var tree []string
	assert.NoError(t, fs.db.Select(&tree, "SELECT layout_path FROM layout_tree WHERE target_path = ?", "a/target2.sub.tmpl"))
	assert.Equal(t, expectTree, tree)
	t.Logf("tree: %v", tree)

	// each target should yield the full tree of templates required to render the template
	expectRenderTree := []string{
		"./_layout.tmpl",
		"a/_sub.layout.tmpl",
		"g/1.tmpl",
		"g/2.tmpl",
		"a/1.tmpl",
		"a/2.tmpl",
	}
	var rTree []string
	assert.NoError(t, fs.db.Select(&rTree, "SELECT layout_path FROM target_tree WHERE target_path = ?", "a/target2.sub.tmpl"))
	assert.Equal(t, expectRenderTree, rTree)
	t.Logf("render tree: %v", rTree)
}
