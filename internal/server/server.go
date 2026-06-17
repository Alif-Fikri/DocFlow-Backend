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

type route struct {
	pattern string
	handler http.HandlerFunc
}

func protectedRoutes(h *handler.Handler) []route {
	return []route{
		{"POST /api/v1/convert", h.Convert},
		{"POST /api/v1/pdf/merge", h.MergePDF},
		{"POST /api/v1/pdf/split", h.SplitPDF},
		{"POST /api/v1/pdf/compress", h.CompressPDF},
	}
}

func (s *Server) Routes() []string {
	h := handler.New(s.cfg, s.conv)
	patterns := []string{"GET /health"}
	for _, r := range protectedRoutes(h) {
		patterns = append(patterns, r.pattern)
	}
	return patterns
}

func (s *Server) Handler() http.Handler {
	h := handler.New(s.cfg, s.conv)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)

	protected := http.NewServeMux()
	for _, r := range protectedRoutes(h) {
		protected.HandleFunc(r.pattern, r.handler)
	}

	mux.Handle("/api/", apiKeyAuth(s.cfg.APIKey, protected))

	return recoverer(logging(mux))
}
