package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Alif-Fikri/DocFlow-Backend/internal/converter"
)

func (h *Handler) Convert(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid or too-large multipart form")
		return
	}

	target := strings.ToLower(strings.TrimSpace(r.FormValue("target")))
	if err := converter.ValidateTarget(target); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}
	file.Close()

	dir, cleanup, err := h.scratchDir("convert-")
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

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "cannot allocate output dir")
		return
	}

	produced, err := h.conv.Convert(r.Context(), inputPath, target, outDir)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	streamFile(w, produced, contentTypeFor(filepath.Ext(produced)))
}
