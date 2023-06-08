package file

import (
	"go-warpdrive/storage"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type files struct {
	dir string
}

func Storage(dir string) storage.File {
	return &files{dir: dir}
}

func (s files) IsExist(err error) bool {
	return os.IsExist(err)
}

func (s files) MkDir(path string, perm os.FileMode) error {
	return os.MkdirAll(s.normalizePath(path), perm)
}

func (s files) FileWriter(path string, perm os.FileMode, offset int64) (w io.WriteCloser, err error) {
	// Try to create storage dir on demand.
	if err = s.MkDir("", 0755); err != nil {
		return
	}
	file, err := os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return
	}
	_, err = file.Seek(offset, 0)
	if err != nil {
		return
	}
	w = file
	return
}

func (s files) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.dir, path)
}
