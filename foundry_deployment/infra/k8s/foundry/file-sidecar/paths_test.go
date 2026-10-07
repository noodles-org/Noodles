package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func newTestRoot(t *testing.T) *root {
	t.Helper()
	dir := t.TempDir()
	j, err := newRoot(dir)
	if err != nil {
		t.Fatalf("newRoot: %v", err)
	}
	return j
}

func TestResolveValidPaths(t *testing.T) {
	j := newTestRoot(t)

	nested := filepath.Join(j.Path, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cases := []struct {
		name string
		rel  string
		want string
	}{
		{"root empty", "", j.Path},
		{"root slash", "/", j.Path},
		{"nested dir", "a/b", nested},
		{"nested file", "a/b/f.txt", filepath.Join(nested, "f.txt")},
		{"dot prefix collapses", "./a/b", nested},
		{"internal dotdot stays inside", "a/b/../b", nested},
		{"nonexistent allowed", "a/new.txt", filepath.Join(j.Path, "a", "new.txt")},
		{"leading slash is root-relative", "/etc/passwd", filepath.Join(j.Path, "etc", "passwd")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := j.resolve(tc.rel)
			if err != nil {
				t.Fatalf("resolve(%q) unexpected error: %v", tc.rel, err)
			}
			if got != tc.want {
				t.Errorf("resolve(%q) = %q, want %q", tc.rel, got, tc.want)
			}
		})
	}
}

func TestResolveRejectsEscapes(t *testing.T) {
	j := newTestRoot(t)

	cases := []struct {
		name string
		rel  string
	}{
		{"parent traversal", "../outside"},
		{"deep traversal", "a/../../outside"},
		{"traversal to etc", "../../../../../../etc/passwd"},
		{"traversal after leading slash", "/../outside"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := j.resolve(tc.rel); err != errEscapes {
				t.Errorf("resolve(%q) error = %v, want errEscapes", tc.rel, err)
			}
		})
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks unreliable on windows")
	}
	j := newTestRoot(t)

	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("s"), 0o644); err != nil {
		t.Fatalf("write outside: %v", err)
	}
	link := filepath.Join(j.Path, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if _, err := j.resolve("escape/secret.txt"); err != errEscapes {
		t.Errorf("resolve through symlink error = %v, want errEscapes", err)
	}
}

func TestNewRootRejectsNonDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := newRoot(file); err == nil {
		t.Error("newRoot on a file should fail")
	}
}

func TestRelPath(t *testing.T) {
	j := newTestRoot(t)
	s := &server{root: j}

	if got := s.relPath(j.Path); got != "" {
		t.Errorf("relPath(root) = %q, want empty", got)
	}
	child := filepath.Join(j.Path, "a", "b.txt")
	if got := s.relPath(child); got != "a/b.txt" {
		t.Errorf("relPath(child) = %q, want a/b.txt", got)
	}
}
