package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alif-Fikri/DocFlow-Backend/internal/config"
	"github.com/Alif-Fikri/DocFlow-Backend/internal/converter"
	"github.com/Alif-Fikri/DocFlow-Backend/internal/server"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the local /health endpoint and exit")
	flag.Parse()

	cfg := config.Load()

	if *healthcheck {
		runHealthcheck(cfg.Port)
		return
	}

	if err := os.MkdirAll(cfg.WorkDir, 0o755); err != nil {
		log.Fatalf("cannot create work dir %q: %v", cfg.WorkDir, err)
	}

	conv := converter.New(cfg)
	srv := server.New(cfg, conv)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	if cfg.APIKey == "" {
		log.Println("WARNING: API_KEY is empty — authentication is DISABLED. Set API_KEY before exposing this service.")
	}
	log.Printf("docflow-backend listening on :%s (soffice=%q, workdir=%q)", cfg.Port, cfg.SofficeBin, cfg.WorkDir)

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}

func runHealthcheck(port string) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil {
		log.Printf("healthcheck failed: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("healthcheck got status %d", resp.StatusCode)
		os.Exit(1)
	}
}
