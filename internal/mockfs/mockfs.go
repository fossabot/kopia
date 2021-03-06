// Package mockfs implements in-memory filesystem for testing.
package mockfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kopia/kopia/fs"
)

// ReaderSeekerCloser implements io.Reader, io.Seeker and io.Closer
type ReaderSeekerCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type readerSeekerCloser struct {
	io.ReadSeeker
}

func (c readerSeekerCloser) Close() error {
	return nil
}

type sortedEntries fs.Entries

func (e sortedEntries) Len() int      { return len(e) }
func (e sortedEntries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e sortedEntries) Less(i, j int) bool {
	return e[i].Metadata().Name < e[j].Metadata().Name
}

type entry struct {
	metadata *fs.EntryMetadata
}

func (ime *entry) Metadata() *fs.EntryMetadata {
	return ime.metadata
}

// Directory is mock in-memory implementation of fs.Directory
type Directory struct {
	entry

	children     fs.Entries
	readdirError error
}

// Summary returns summary of a directory.
func (imd *Directory) Summary() *fs.DirectorySummary {
	return nil
}

// AddFileLines adds a mock file with the specified name, text content and permissions.
func (imd *Directory) AddFileLines(name string, lines []string, permissions fs.Permissions) *File {
	return imd.AddFile(name, []byte(strings.Join(lines, "\n")), permissions)
}

// AddFile adds a mock file with the specified name, content and permissions.
func (imd *Directory) AddFile(name string, content []byte, permissions fs.Permissions) *File {
	imd, name = imd.resolveSubdir(name)
	file := &File{
		entry: entry{
			metadata: &fs.EntryMetadata{
				Name:        name,
				Type:        fs.EntryTypeFile,
				Permissions: permissions,
				FileSize:    int64(len(content)),
			},
		},
		source: func() (ReaderSeekerCloser, error) {
			return readerSeekerCloser{bytes.NewReader(content)}, nil
		},
	}

	imd.addChild(file)

	return file
}

// AddDir adds a fake directory with a given name and permissions.
func (imd *Directory) AddDir(name string, permissions fs.Permissions) *Directory {
	imd, name = imd.resolveSubdir(name)

	subdir := &Directory{
		entry: entry{
			metadata: &fs.EntryMetadata{
				Name:        name,
				Type:        fs.EntryTypeDirectory,
				Permissions: permissions,
			},
		},
	}

	imd.addChild(subdir)

	return subdir
}

func (imd *Directory) addChild(e fs.Entry) {
	if strings.Contains(e.Metadata().Name, "/") {
		panic("child name cannot contain '/'")
	}
	imd.children = append(imd.children, e)
	sort.Sort(sortedEntries(imd.children))
}

func (imd *Directory) resolveSubdir(name string) (*Directory, string) {
	parts := strings.Split(name, "/")
	for _, n := range parts[0 : len(parts)-1] {
		imd = imd.Subdir(n)
	}
	return imd, parts[len(parts)-1]
}

// Subdir finds a subdirectory with a given name.
func (imd *Directory) Subdir(name ...string) *Directory {
	i := imd
	for _, n := range name {
		i2 := i.children.FindByName(n)
		if i2 == nil {
			panic(fmt.Sprintf("'%s' not found in '%s'", n, i.metadata.Name))
		}
		if !i2.Metadata().FileMode().IsDir() {
			panic(fmt.Sprintf("'%s' is not a directory in '%s'", n, i.metadata.Name))
		}

		i = i2.(*Directory)
	}
	return i
}

// Remove removes directory entry with a given name.
func (imd *Directory) Remove(name string) {
	newChildren := imd.children[:0]

	for _, e := range imd.children {
		if e.Metadata().Name != name {
			newChildren = append(newChildren, e)
		}
	}

	imd.children = newChildren
}

// FailReaddir causes the subsequent Readdir() calls to fail with the specified error.
func (imd *Directory) FailReaddir(err error) {
	imd.readdirError = err
}

// Readdir gets the contents of a directory.
func (imd *Directory) Readdir(ctx context.Context) (fs.Entries, error) {
	if imd.readdirError != nil {
		return nil, imd.readdirError
	}

	return append(fs.Entries(nil), imd.children...), nil
}

// File is an in-memory fs.File capable of simulating failures.
type File struct {
	entry

	source func() (ReaderSeekerCloser, error)
}

// SetContents changes the contents of a given file.
func (imf *File) SetContents(b []byte) {
	imf.source = func() (ReaderSeekerCloser, error) {
		return readerSeekerCloser{bytes.NewReader(b)}, nil
	}

}

type fileReader struct {
	ReaderSeekerCloser
	metadata *fs.EntryMetadata
}

func (ifr *fileReader) EntryMetadata() (*fs.EntryMetadata, error) {
	return ifr.metadata, nil
}

// Open opens the file for reading, optionally simulating error.
func (imf *File) Open(ctx context.Context) (fs.Reader, error) {
	r, err := imf.source()
	if err != nil {
		return nil, err
	}

	return &fileReader{
		ReaderSeekerCloser: r,
		metadata:           imf.metadata,
	}, nil
}

type inmemorySymlink struct {
	entry
}

func (imsl *inmemorySymlink) Readlink(ctx context.Context) (string, error) {
	panic("not implemented yet")
}

// NewDirectory returns new mock directory.ds
func NewDirectory() *Directory {
	return &Directory{
		entry: entry{
			metadata: &fs.EntryMetadata{
				Name: "<root>",
			},
		},
	}
}

var _ fs.Directory = &Directory{}
var _ fs.File = &File{}
var _ fs.Symlink = &inmemorySymlink{}
