// Command server starts the SDM & Legal "Legal-Permit War Room" API.
//
// It wires the layers together — repository -> service -> HTTP transport — and
// runs an HTTP server with graceful shutdown.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"greenpark/sdm/internal/auth"
	"greenpark/sdm/internal/config"
	"greenpark/sdm/internal/repository"
	"greenpark/sdm/internal/service"
	httptransport "greenpark/sdm/internal/transport/http"
)

func main() {
	cfg := config.Load()

	repo, err := repository.NewRepository(cfg.DataPath)
	if err != nil {
		log.Fatalf("sdm: failed to open data store %q: %v", cfg.DataPath, err)
	}
	authSvc := auth.New(repo, cfg.SessionTTL)
	svc := service.New(repo, authSvc)
	handler := httptransport.NewHandler(svc)
	router := httptransport.NewRouter(handler, cfg.AllowOrigin)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("sdm API listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("sdm: server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("sdm: shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("sdm: graceful shutdown failed: %v", err)
	}
	log.Println("sdm: stopped")
}
