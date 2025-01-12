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
	"notifyHub/internal/handler"
	"notifyHub/internal/queue"
	"notifyHub/internal/repository"
	"notifyHub/internal/service"
	"notifyHub/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := repository.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	producer := queue.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error("kafka producer close failed", "error", err)
		}
	}()

	svc := service.NewNotificationService(service.Deps{
		Repo:     repository.NewNotificationRepository(pool),
		Logs:     repository.NewDeliveryLogRepository(pool),
		Producer: producer,
	})
	h := handler.NewNotificationHandler(svc)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler.NewRouter(h),
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("api shutdown error", "error", err)
		os.Exit(1)
	}

	fmt.Println()
	log.Info("api stopped")
}
