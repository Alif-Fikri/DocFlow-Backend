package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (h *Handler) MergePDF(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid or too-large multipart form")
		return
	}

	headers := r.MultipartForm.File["files"]
	if len(headers) < 2 {
		writeError(w, http.StatusBadRequest, "merge requires at least 2 'files'")
		return
	}

	dir, cleanup, err := h.scratchDir("merge-")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot allocate workspace")
		return
	}
	defer cleanup()

	var inputs []string
	for _, fh := range headers {
		path, err := saveUpload(fh, dir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "cannot save upload")
			return
		}
		inputs = append(inputs, path)
	}

	outPath := filepath.Join(dir, "merged.pdf")
	if err := h.conv.MergePDF(inputs, outPath); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	streamFile(w, outPath, "application/pdf")
}

func (h *Handler) SplitPDF(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid or too-large multipart form")
		return
	}

	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}

	dir, cleanup, err := h.scratchDir("split-")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot allocate workspace")
		return
	}
	defer cleanup()

	inputPath, err := saveUpload(fileHeader, dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot save upload")
		return
	}

	outDir := filepath.Join(dir, "pages")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "cannot allocate output dir")
		return
	}

	parts, err := h.conv.SplitPDF(inputPath, outDir)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	zipPath := filepath.Join(dir, base+"_pages.zip")
	if err := zipFiles(parts, zipPath); err != nil {
		writeError(w, http.StatusInternalServerError, "cannot build zip")
		return
	}
	streamFile(w, zipPath, "application/zip")
}

func (h *Handler) CompressPDF(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid or too-large multipart form")
		return
	}

	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}

	dir, cleanup, err := h.scratchDir("compress-")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot allocate workspace")
		return
	}
	defer cleanup()

	inputPath, err := saveUpload(fileHeader, dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot save upload")
		return
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outPath := filepath.Join(dir, base+"_compressed.pdf")
	if err := h.conv.CompressPDF(inputPath, outPath); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	streamFile(w, outPath, "application/pdf")
}
