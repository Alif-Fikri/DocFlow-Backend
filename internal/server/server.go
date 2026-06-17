package server

import (
	"net/http"

	"github.com/Alif-Fikri/DocFlow-Backend/internal/config"
	"github.com/Alif-Fikri/DocFlow-Backend/internal/converter"
	"github.com/Alif-Fikri/DocFlow-Backend/internal/handler"
)

type Server struct {
	cfg  config.Config
	conv *converter.Converter
}

func New(cfg config.Config, conv *converter.Converter) *Server {
	return &Server{cfg: cfg, conv: conv}
}

func (s *Server) Handler() http.Handler {
	h := handler.New(s.cfg, s.conv)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)

	protected := http.NewServeMux()
	protected.HandleFunc("POST /api/v1/convert", h.Convert)
	protected.HandleFunc("POST /api/v1/pdf/merge", h.MergePDF)
	protected.HandleFunc("POST /api/v1/pdf/split", h.SplitPDF)
	protected.HandleFunc("POST /api/v1/pdf/compress", h.CompressPDF)

	mux.Handle("/api/", apiKeyAuth(s.cfg.APIKey, protected))

	return recoverer(logging(mux))
}
