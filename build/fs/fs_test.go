package fs

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskBackedRead(t *testing.T) {
	td, err := os.MkdirTemp("", "taevas")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	testData := []byte("this is a test")
	path := filepath.Join(td, "test.dat")
	if err := ioutil.WriteFile(path, testData, fs.ModePerm); err != nil {
		require.NoError(t, err)
	}

	fs, err := New(td)
	require.NoError(t, err)

	if _, err := fs.Add("test.dat"); err != nil {
		require.NoError(t, err)
	}

	// it should read and stat a disk-backed file
	f, err := fs.Open("test.dat")
	require.NoError(t, err)
	finfo, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testData)), finfo.Size())
	data, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// it should get directory entries
	entries, err := fs.ReadDir(".")
	require.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "test.dat", entries[0].Name())
	finfo2, err := entries[0].Info()
	require.NoError(t, err)
	assert.Equal(t, int64(len(testData)), finfo2.Size())

	// it should read files
	b1, err := fs.ReadFile("test.dat")
	require.NoError(t, err)
	assert.Equal(t, testData, b1)
	b2, err := fs.ReadFile("./test.dat")
	require.NoError(t, err)
	assert.Equal(t, testData, b2)
}

func TestVirtualRead(t *testing.T) {
	testData := []byte("this is a test")

	fs, err := New("/test")
	require.NoError(t, err)

	if _, err := fs.AddVirtual("test.dat", testData); err != nil {
		require.NoError(t, err)
	}

	// it should read and stat a virtual file
	f, err := fs.Open("test.dat")
	require.NoError(t, err)
	finfo, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testData)), finfo.Size())
	data, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, testData, data)

	// it should get directory entries
	entries, err := fs.ReadDir(".")
	require.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "test.dat", entries[0].Name())
	finfo2, err := entries[0].Info()
	require.NoError(t, err)
	assert.Equal(t, int64(len(testData)), finfo2.Size())

	// it should read files
	b1, err := fs.ReadFile("test.dat")
	require.NoError(t, err)
	assert.Equal(t, testData, b1)
	b2, err := fs.ReadFile("./test.dat")
	require.NoError(t, err)
	assert.Equal(t, testData, b2)
}

func TestFlush(t *testing.T) {
	td, err := os.MkdirTemp("", "taevas")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	testData := []byte("this is a test")

	fs, err := New(td)
	require.NoError(t, err)
	if _, err := fs.AddVirtual("test.dat", testData); err != nil {
		require.NoError(t, err)
	}

	require.NoError(t, fs.Flush())

	got, err := os.ReadFile(filepath.Join(td, "test.dat"))
	require.NoError(t, err)
	assert.Equal(t, testData, got)

}

//go:embed testdata
var fsTruth embed.FS

func TestReadDirSort(t *testing.T) {
	td, err := os.MkdirTemp("", "taevas")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	dbPath := filepath.Join(td, "test.db")

	paths := []string{
		"1.dat",
		"2.dat",
		"a.dat",
		"b.dat",
		"a/1.dat",
		"a/2.dat",
		"a/a.dat",
		"b/1.dat",
		"b/2.dat",
		"b/a.dat",
		"b/b.dat",
		"b/c/1.dat",
	}

	fs, err := New("testdata", WithDBFile(dbPath))
	require.NoError(t, err)

	for _, p := range paths {
		if _, err := fs.AddVirtual(p, []byte("test")); err != nil {
			require.NoError(t, err)
		}
	}

	dTruth, err := fsTruth.ReadDir("testdata")
	require.NoError(t, err)

	dExperimental, err := fs.ReadDir(".")
	require.NoError(t, err)

	trueSort := make([]string, len(dTruth))
	for i, e := range dTruth {
		trueSort[i] = e.Name()
	}

	expSort := make([]string, len(dExperimental))
	for i, e := range dExperimental {
		expSort[i] = e.Name()
	}

	require.Equal(t, len(trueSort), len(expSort))
	assert.Equal(t, trueSort, expSort)
}
