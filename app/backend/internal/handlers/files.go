package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/mephalrith/noodles/backend/internal/errs"
	"github.com/mephalrith/noodles/backend/internal/middleware"
	"github.com/mephalrith/noodles/backend/internal/respond"
	"github.com/mephalrith/noodles/backend/internal/services"
)

const maxUploadSize = 2 << 30

func validPath(p string) bool {
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") {
		return false
	}
	return !slices.Contains(strings.Split(p, "/"), "..")
}

func recordFileAction(action, status string) {
	services.FileActions.With(map[string]string{"action": action, "status": status}).Inc()
}

func HandleListFiles(files *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rel := r.URL.Query().Get("path")
		if !validPath(rel) {
			respond.Error(w, errs.InvalidPath)
			return
		}

		entries, err := files.List(r.Context(), rel)
		if err != nil {
			respond.Error(w, asError(err))
			return
		}

		respond.OK(w, entries)
	}
}

func HandleDownloadFile(files *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rel := r.URL.Query().Get("path")
		if rel == "" {
			respond.Error(w, errs.MissingPath)
			return
		}
		if !validPath(rel) {
			respond.Error(w, errs.InvalidPath)
			return
		}

		dl, err := files.Download(r.Context(), rel)
		if err != nil {
			respond.Error(w, asError(err))
			return
		}
		defer dl.Body.Close()

		w.Header().Set("Content-Disposition", "attachment; filename=\""+dl.Name+"\"")
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := io.Copy(w, dl.Body); err != nil {
			services.Logger.Error("File download stream failed", "path", rel, "error", err)
		}
	}
}

func HandleUploadFile(files *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("path")
		if !validPath(dir) {
			respond.Error(w, errs.InvalidPath)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		file, header, err := r.FormFile("file")
		if err != nil {
			respond.Error(w, errs.BadRequest)
			return
		}
		defer file.Close()

		name := path.Base(header.Filename)
		if err := files.Upload(r.Context(), dir, name, file); err != nil {
			recordFileAction("upload", "error")
			respond.Error(w, asError(err))
			return
		}

		user := middleware.UserFromContext(r.Context())
		recordFileAction("upload", "ok")
		services.Logger.Info("File uploaded", "dir", dir, "name", name, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func HandleDeleteFile(files *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rel := r.URL.Query().Get("path")
		if rel == "" {
			respond.Error(w, errs.MissingPath)
			return
		}
		if !validPath(rel) {
			respond.Error(w, errs.InvalidPath)
			return
		}

		if err := files.Delete(r.Context(), rel); err != nil {
			recordFileAction("delete", "error")
			respond.Error(w, asError(err))
			return
		}

		user := middleware.UserFromContext(r.Context())
		recordFileAction("delete", "ok")
		services.Logger.Info("File deleted", "path", rel, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func HandleMkdir(files *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rel := r.URL.Query().Get("path")
		if rel == "" {
			respond.Error(w, errs.MissingPath)
			return
		}
		if !validPath(rel) {
			respond.Error(w, errs.InvalidPath)
			return
		}

		if err := files.Mkdir(r.Context(), rel); err != nil {
			recordFileAction("mkdir", "error")
			respond.Error(w, asError(err))
			return
		}

		user := middleware.UserFromContext(r.Context())
		recordFileAction("mkdir", "ok")
		services.Logger.Info("Directory created", "path", rel, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func HandleRename(files *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.Error(w, errs.BadRequest)
			return
		}
		if !validPath(req.From) || !validPath(req.To) || req.From == "" || req.To == "" {
			respond.Error(w, errs.InvalidPath)
			return
		}

		if err := files.Rename(r.Context(), req.From, req.To); err != nil {
			recordFileAction("rename", "error")
			respond.Error(w, asError(err))
			return
		}

		user := middleware.UserFromContext(r.Context())
		recordFileAction("rename", "ok")
		services.Logger.Info("File renamed", "from", req.From, "to", req.To, "user", user.Email)

		respond.OK(w, map[string]bool{"ok": true})
	}
}

func asError(err error) *errs.Error {
	var e *errs.Error
	if errors.As(err, &e) {
		return e
	}
	return errs.Internal("File operation failed")
}
