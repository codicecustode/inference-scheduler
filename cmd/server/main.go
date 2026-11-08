package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codicecustode/inference-scheduler/config"
	"github.com/codicecustode/inference-scheduler/internal/api"
	"github.com/codicecustode/inference-scheduler/internal/models"
	"github.com/codicecustode/inference-scheduler/internal/scheduler"
	"github.com/codicecustode/inference-scheduler/internal/store"
	"github.com/codicecustode/inference-scheduler/internal/stream"
)

func main() {
	//Load config
	cfg := config.Load()

	//Create store
	st := store.NewMemoryStore()

	//Create event channel (shared between scheduler and broker)
	eventCh := make(chan *models.Event, 100)

	//Create worker pool and start it
	pool := scheduler.NewWorkerPool(cfg.NumGPUs, cfg.VRAMPerGPU)
	pool.Start()

	//Create scheduler and start it
	sched := scheduler.NewScheduler(st, pool, eventCh)
	sched.Start()

	//Create SSE broker and start it
	broker := stream.NewBroker(eventCh)
	broker.Start()

	//Create handler
	handler := api.NewHandler(sched, st, pool)

	//Create router with middleware
	router := api.NewRouter(handler, broker, cfg.APIKey, cfg.RateLimit)

	//Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	//Start server in a background goroutine
	go func() {
		log.Printf("server starting on port %d", cfg.Port)
		log.Printf("GPUs: %d, VRAM per GPU: %dMB", cfg.NumGPUs, cfg.VRAMPerGPU)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	//Wait for Ctrl+C (SIGINT) or kill (SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	//Graceful shutdown — give 10 seconds for in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	sched.Stop()
	pool.Stop()

	log.Println("server stopped")
}
