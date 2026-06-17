package handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Alif-Fikri/DocFlow-Backend/internal/config"
	"github.com/Alif-Fikri/DocFlow-Backend/internal/converter"
)

type Handler struct {
	cfg  config.Config
	conv *converter.Converter
}

func New(cfg config.Config, conv *converter.Converter) *Handler {
	return &Handler{cfg: cfg, conv: conv}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) scratchDir(prefix string) (string, func(), error) {
	dir, err := os.MkdirTemp(h.cfg.WorkDir, prefix)
	if err != nil {
		return "", nil, err
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}

func saveUpload(fh *multipart.FileHeader, dir string) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	name := filepath.Base(fh.Filename)
	if name == "." || name == "/" || name == "" {
		name = "upload"
	}
	dst := filepath.Join(dir, name)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return dst, nil
}

func streamFile(w http.ResponseWriter, path, contentType string) {
	f, err := os.Open(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot open result file")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot stat result file")
		return
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(path)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	_, _ = io.Copy(w, f)
}

func zipFiles(files []string, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	for _, file := range files {
		if err := addToZip(zw, file); err != nil {
			return err
		}
	}
	return nil
}

func addToZip(zw *zip.Writer, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := zw.Create(filepath.Base(file))
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func contentTypeFor(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".html":
		return "text/html"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
