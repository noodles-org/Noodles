package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *server {
	t.Helper()
	return &server{root: newTestRoot(t)}
}

func TestHandleList(t *testing.T) {
	s := newTestServer(t)
	if err := os.Mkdir(filepath.Join(s.root.Path, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.root.Path, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/files?path=", nil)
	rec := httptest.NewRecorder()
	s.handleList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var entries []fileEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if !entries[0].IsDir || entries[0].Name != "sub" {
		t.Errorf("first entry = %+v, want dir 'sub'", entries[0])
	}
	if entries[1].Name != "a.txt" || entries[1].Size != 2 {
		t.Errorf("second entry = %+v, want a.txt size 2", entries[1])
	}
}

func TestHandleDeleteFile(t *testing.T) {
	s := newTestServer(t)
	target := filepath.Join(s.root.Path, "dup.png")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/files?path=dup.png", nil)
	rec := httptest.NewRecorder()
	s.handleDelete(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestHandleDeleteEmptyDir(t *testing.T) {
	s := newTestServer(t)
	dir := filepath.Join(s.root.Path, "empty")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/files?path=empty", nil)
	rec := httptest.NewRecorder()
	s.handleDelete(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("empty dir should have been deleted")
	}
}

func TestHandleDeleteNonEmptyDir(t *testing.T) {
	s := newTestServer(t)
	dir := filepath.Join(s.root.Path, "full")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/files?path=full", nil)
	rec := httptest.NewRecorder()
	s.handleDelete(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Error("non-empty dir should not have been deleted")
	}
}

func TestHandleDeleteRejectsEscape(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodDelete, "/files?path=../../etc/hosts", nil)
	rec := httptest.NewRecorder()
	s.handleDelete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUpload(t *testing.T) {
	s := newTestServer(t)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "image.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("data")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload?path=", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	s.handleUpload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	got, err := os.ReadFile(filepath.Join(s.root.Path, "image.png"))
	if err != nil {
		t.Fatalf("read uploaded: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("uploaded content = %q, want data", got)
	}
}

func TestHandleMkdirAndRename(t *testing.T) {
	s := newTestServer(t)

	mkReq := httptest.NewRequest(http.MethodPost, "/mkdir?path=worlds", nil)
	mkRec := httptest.NewRecorder()
	s.handleMkdir(mkRec, mkReq)
	if mkRec.Code != http.StatusOK {
		t.Fatalf("mkdir status = %d, want 200", mkRec.Code)
	}
	if info, err := os.Stat(filepath.Join(s.root.Path, "worlds")); err != nil || !info.IsDir() {
		t.Fatalf("worlds dir not created: %v", err)
	}

	renReq := httptest.NewRequest(http.MethodPost, "/rename",
		strings.NewReader(`{"from":"worlds","to":"campaigns"}`))
	renRec := httptest.NewRecorder()
	s.handleRename(renRec, renReq)
	if renRec.Code != http.StatusOK {
		t.Fatalf("rename status = %d, want 200 (%s)", renRec.Code, renRec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(s.root.Path, "campaigns")); err != nil {
		t.Errorf("renamed dir missing: %v", err)
	}
}

func TestRequireToken(t *testing.T) {
	called := false
	h := requireToken("secret", func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/files", nil))
	if rec.Code != http.StatusUnauthorized || called {
		t.Fatalf("missing token: status=%d called=%v", rec.Code, called)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	req.Header.Set("Authorization", "Bearer secret")
	h(rec, req)
	if rec.Code != http.StatusOK || !called {
		t.Fatalf("valid token: status=%d called=%v", rec.Code, called)
	}
}
