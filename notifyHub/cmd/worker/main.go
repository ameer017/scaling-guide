package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"notifyHub/internal/config"
	"notifyHub/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	log.Info("worker starting",
		"env", cfg.AppEnv,
		"kafka_brokers", cfg.KafkaBrokers,
		"kafka_topic", cfg.KafkaTopic,
		"kafka_group_id", cfg.KafkaGroupID,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Placeholder loop — Kafka consumer lands in a later step.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopped")
			return
		case <-ticker.C:
			log.Info("worker idle — waiting for Kafka consumer implementation",
				"topic", cfg.KafkaTopic,
			)
		}
	}
}
