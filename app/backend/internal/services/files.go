package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/model"
)

type Download struct {
	Body    io.ReadCloser
	Name    string
	Size    int64
	ModTime time.Time
}

type FileService struct {
	baseURL string
	token   string
	client  *http.Client
	isDev   bool
	devRoot string
}

func NewFileService(cfg *config.Config) *FileService {
	svc := &FileService{
		baseURL: strings.TrimRight(cfg.FileSvc.URL, "/"),
		token:   cfg.FileSvc.Token,
		client:  &http.Client{Timeout: 30 * time.Second},
		isDev:   !cfg.IsProduction,
		devRoot: cfg.FileSvc.Root,
	}

	if svc.isDev {
		seed := svc.devRoot
		if abs, err := filepath.Abs(seed); err == nil {
			seed = abs
		}

		work, err := os.MkdirTemp(filepath.Dir(seed), "files-work-")
		if err != nil {
			Logger.Error("Files: failed to create dev work dir", "error", err)
			svc.devRoot = seed
			return svc
		}

		if err := copyTree(seed, work); err != nil {
			Logger.Error("Files: failed to seed dev work dir", "seed", seed, "work", work, "error", err)
		} else {
			Logger.Info("Files: dev mode, serving from ephemeral copy", "seed", seed, "work", work)
		}
		svc.devRoot = work
	}

	return svc
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

func (s *FileService) List(ctx context.Context, rel string) ([]model.FileEntry, error) {
	if s.isDev {
		return s.listLocal(rel)
	}

	resp, err := s.do(ctx, http.MethodGet, "/files", url.Values{"path": {rel}}, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if e := statusError(resp); e != nil {
		return nil, e
	}

	var entries []model.FileEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, errs.Internal("Failed to decode file list")
	}
	return entries, nil
}

func (s *FileService) Download(ctx context.Context, rel string) (*Download, error) {
	if s.isDev {
		return s.downloadLocal(rel)
	}

	resp, err := s.do(ctx, http.MethodGet, "/download", url.Values{"path": {rel}}, nil, "")
	if err != nil {
		return nil, err
	}

	if e := statusError(resp); e != nil {
		resp.Body.Close()
		return nil, e
	}

	name := path.Base(rel)
	if name == "" || name == "." || name == "/" {
		name = "download"
	}
	return &Download{Body: resp.Body, Name: name}, nil
}

func (s *FileService) Upload(ctx context.Context, dir, name string, src io.Reader) error {
	if s.isDev {
		return s.uploadLocal(dir, name, src)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return errs.Internal("Failed to build upload")
	}
	if _, err := io.Copy(part, src); err != nil {
		return errs.Internal("Failed to read upload")
	}
	if err := writer.Close(); err != nil {
		return errs.Internal("Failed to build upload")
	}

	resp, err := s.do(ctx, http.MethodPost, "/upload", url.Values{"path": {dir}}, &body, writer.FormDataContentType())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return statusError(resp)
}

func (s *FileService) Delete(ctx context.Context, rel string) error {
	if s.isDev {
		return s.deleteLocal(rel)
	}

	resp, err := s.do(ctx, http.MethodDelete, "/files", url.Values{"path": {rel}}, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return statusError(resp)
}

func (s *FileService) Mkdir(ctx context.Context, rel string) error {
	if s.isDev {
		return s.mkdirLocal(rel)
	}

	resp, err := s.do(ctx, http.MethodPost, "/mkdir", url.Values{"path": {rel}}, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return statusError(resp)
}

func (s *FileService) Rename(ctx context.Context, from, to string) error {
	if s.isDev {
		return s.renameLocal(from, to)
	}

	payload, err := json.Marshal(map[string]string{"from": from, "to": to})
	if err != nil {
		return errs.Internal("Failed to build rename")
	}

	resp, err := s.do(ctx, http.MethodPost, "/rename", nil, bytes.NewReader(payload), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return statusError(resp)
}

func (s *FileService) do(ctx context.Context, method, endpoint string, query url.Values, body io.Reader, contentType string) (*http.Response, error) {
	u := s.baseURL + endpoint
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, errs.Internal("Failed to reach file service")
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		Logger.Error("Files: sidecar request failed", "endpoint", endpoint, "error", err)
		return nil, errs.Internal("File service unavailable")
	}
	return resp, nil
}

func statusError(resp *http.Response) *errs.Error {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return errs.FileNotFound
	case resp.StatusCode == http.StatusConflict:
		return errs.DirNotEmpty
	case resp.StatusCode == http.StatusBadRequest:
		return errs.InvalidPath
	default:
		return errs.Internal("File operation failed")
	}
}

func (s *FileService) resolve(rel string) (string, error) {
	clean := filepath.Clean("/" + filepath.FromSlash(rel))
	abs := filepath.Join(s.devRoot, clean)
	if abs != s.devRoot && !strings.HasPrefix(abs, s.devRoot+string(os.PathSeparator)) {
		return "", errs.InvalidPath
	}
	return abs, nil
}

func (s *FileService) relPath(abs string) string {
	rel, err := filepath.Rel(s.devRoot, abs)
	if err != nil || rel == "." {
		return ""
	}
	return filepath.ToSlash(rel)
}

func (s *FileService) listLocal(rel string) ([]model.FileEntry, error) {
	abs, err := s.resolve(rel)
	if err != nil {
		return nil, errs.InvalidPath
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errs.FileNotFound
		}
		return nil, errs.Internal("Failed to list files")
	}

	out := make([]model.FileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		child := filepath.Join(abs, e.Name())
		out = append(out, model.FileEntry{
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
	return out, nil
}

func (s *FileService) downloadLocal(rel string) (*Download, error) {
	abs, err := s.resolve(rel)
	if err != nil {
		return nil, errs.InvalidPath
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, errs.FileNotFound
	}
	if info.IsDir() {
		return nil, errs.InvalidPath
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, errs.Internal("Failed to open file")
	}
	return &Download{Body: f, Name: info.Name(), Size: info.Size(), ModTime: info.ModTime()}, nil
}

func (s *FileService) uploadLocal(dir, name string, src io.Reader) error {
	abs, err := s.resolve(dir)
	if err != nil {
		return errs.InvalidPath
	}

	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return errs.InvalidPath
	}

	clean := filepath.Base(filepath.FromSlash(name))
	if clean == "." || clean == string(os.PathSeparator) || clean == "" {
		return errs.InvalidPath
	}

	dest, err := s.resolve(path.Join(s.relPath(abs), clean))
	if err != nil {
		return errs.InvalidPath
	}

	out, err := os.Create(dest)
	if err != nil {
		return errs.Internal("Failed to write file")
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return errs.Internal("Failed to write file")
	}
	return nil
}

func (s *FileService) deleteLocal(rel string) error {
	abs, err := s.resolve(rel)
	if err != nil {
		return errs.InvalidPath
	}
	if abs == s.devRoot {
		return errs.InvalidPath
	}
	if _, err := os.Stat(abs); err != nil {
		return errs.FileNotFound
	}
	if err := os.Remove(abs); err != nil {
		if errors.Is(err, syscall.ENOTEMPTY) {
			return errs.DirNotEmpty
		}
		return errs.Internal("Failed to delete")
	}
	return nil
}

func (s *FileService) mkdirLocal(rel string) error {
	abs, err := s.resolve(rel)
	if err != nil {
		return errs.InvalidPath
	}
	if abs == s.devRoot {
		return errs.InvalidPath
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return errs.Internal("Failed to create directory")
	}
	return nil
}

func (s *FileService) renameLocal(from, to string) error {
	fromAbs, err := s.resolve(from)
	if err != nil {
		return errs.InvalidPath
	}
	toAbs, err := s.resolve(to)
	if err != nil {
		return errs.InvalidPath
	}
	if fromAbs == s.devRoot || toAbs == s.devRoot {
		return errs.InvalidPath
	}
	if _, err := os.Stat(fromAbs); err != nil {
		return errs.FileNotFound
	}
	if err := os.Rename(fromAbs, toAbs); err != nil {
		return errs.Internal("Failed to rename")
	}
	return nil
}
