package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notifyHub/internal/config"
	"notifyHub/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"notifyhub-api"}`))
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("api starting",
			"port", cfg.HTTPPort,
			"env", cfg.AppEnv,
			"kafka_brokers", cfg.KafkaBrokers,
			"kafka_topic", cfg.KafkaTopic,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("api shutdown error", "error", err)
		os.Exit(1)
	}

	fmt.Println()
	log.Info("api stopped")
}
