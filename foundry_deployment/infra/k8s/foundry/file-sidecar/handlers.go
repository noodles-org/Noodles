package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type fileEntry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

type renameRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type server struct {
	root *root
}

func (s *server) relPath(abs string) string {
	rel, err := filepath.Rel(s.root.Path, abs)
	if err != nil || rel == "." {
		return ""
	}
	return filepath.ToSlash(rel)
}

func (s *server) resolveParam(w http.ResponseWriter, r *http.Request, param string) (string, bool) {
	abs, err := s.root.resolve(r.URL.Query().Get(param))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid path")
		return "", false
	}
	return abs, true
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	abs, ok := s.resolveParam(w, r, "path")
	if !ok {
		return
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		log.Printf("list %q: %v", abs, err)
		writeError(w, http.StatusInternalServerError, "list failed")
		return
	}

	out := make([]fileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		child := filepath.Join(abs, e.Name())
		out = append(out, fileEntry{
			Name:    e.Name(),
			Path:    s.relPath(child),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().UTC(),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})

	writeJSON(w, http.StatusOK, out)
}

func (s *server) handleDownload(w http.ResponseWriter, r *http.Request) {
	abs, ok := s.resolveParam(w, r, "path")
	if !ok {
		return
	}

	info, err := os.Stat(abs)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "cannot download a directory")
		return
	}

	f, err := os.Open(abs)
	if err != nil {
		log.Printf("download %q: %v", abs, err)
		writeError(w, http.StatusInternalServerError, "download failed")
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("close %q: %v", abs, cerr)
		}
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(abs)+"\"")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	dir, ok := s.resolveParam(w, r, "path")
	if !ok {
		return
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		writeError(w, http.StatusBadRequest, "target directory does not exist")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("close upload: %v", cerr)
		}
	}()

	name := filepath.Base(filepath.FromSlash(header.Filename))
	if name == "." || name == string(os.PathSeparator) || name == "" {
		writeError(w, http.StatusBadRequest, "invalid file name")
		return
	}
	dest, err := s.root.resolve(path.Join(s.relPath(dir), name))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	out, err := os.Create(dest)
	if err != nil {
		log.Printf("create %q: %v", dest, err)
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			log.Printf("close %q: %v", dest, cerr)
		}
	}()

	if _, err := io.Copy(out, file); err != nil {
		log.Printf("write %q: %v", dest, err)
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"path": s.relPath(dest)})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	abs, ok := s.resolveParam(w, r, "path")
	if !ok {
		return
	}

	if abs == s.root.Path {
		writeError(w, http.StatusBadRequest, "cannot delete root")
		return
	}

	if _, err := os.Stat(abs); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := os.Remove(abs); err != nil {
		if errors.Is(err, syscall.ENOTEMPTY) {
			writeError(w, http.StatusConflict, "directory not empty")
			return
		}
		log.Printf("delete %q: %v", abs, err)
		writeError(w, http.StatusInternalServerError, "delete failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *server) handleMkdir(w http.ResponseWriter, r *http.Request) {
	abs, ok := s.resolveParam(w, r, "path")
	if !ok {
		return
	}

	if abs == s.root.Path {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	if err := os.MkdirAll(abs, 0o755); err != nil {
		log.Printf("mkdir %q: %v", abs, err)
		writeError(w, http.StatusInternalServerError, "mkdir failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"path": s.relPath(abs)})
}

func (s *server) handleRename(w http.ResponseWriter, r *http.Request) {
	var req renameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	from, err := s.root.resolve(req.From)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid source path")
		return
	}
	to, err := s.root.resolve(req.To)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid destination path")
		return
	}
	if from == s.root.Path || to == s.root.Path {
		writeError(w, http.StatusBadRequest, "cannot rename root")
		return
	}

	if _, err := os.Stat(from); err != nil {
		writeError(w, http.StatusNotFound, "source not found")
		return
	}

	if err := os.Rename(from, to); err != nil {
		log.Printf("rename %q -> %q: %v", from, to, err)
		writeError(w, http.StatusInternalServerError, "rename failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"path": s.relPath(to)})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
