package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var errEscapes = errors.New("path escapes root")

type root struct {
	Path string
}

func newRoot(dir string) (*root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("root is not a directory")
	}

	return &root{Path: resolved}, nil
}

func (j *root) resolve(rel string) (string, error) {
	slash := strings.ReplaceAll(rel, "\\", "/")
	trimmed := strings.TrimLeft(slash, "/")

	abs := filepath.Join(j.Path, filepath.FromSlash(trimmed))

	if !withinRoot(j.Path, abs) {
		return "", errEscapes
	}
	if err := j.checkSymlinks(abs); err != nil {
		return "", err
	}

	return abs, nil
}

func (j *root) checkSymlinks(abs string) error {
	probe := abs
	for {
		resolved, err := filepath.EvalSymlinks(probe)
		if err == nil {
			if !withinRoot(j.Path, resolved) {
				return errEscapes
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			// Reached the filesystem root without finding an existing ancestor.
			return nil
		}
		probe = parent
	}
}

func withinRoot(root, path string) bool {
	if path == root {
		return true
	}
	return strings.HasPrefix(path, root+string(os.PathSeparator))
}
