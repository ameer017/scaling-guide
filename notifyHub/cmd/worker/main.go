package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notifyHub/internal/config"
	"notifyHub/internal/queue"
	"notifyHub/internal/repository"
	"notifyHub/internal/service"
	"notifyHub/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Worker only needs the repo for status updates; producer is unused.
	repo := repository.NewNotificationRepository(pool)
	svc := service.NewNotificationService(repo, nil)

	consumer := queue.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error("kafka consumer close failed", "error", err)
		}
	}()

	log.Info("worker starting",
		"env", cfg.AppEnv,
		"kafka_brokers", cfg.KafkaBrokers,
		"kafka_topic", cfg.KafkaTopic,
		"kafka_group_id", cfg.KafkaGroupID,
	)

	for {
		msg, payload, err := consumer.Fetch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
				log.Info("worker stopped")
				return
			}
			log.Error("kafka fetch failed", "error", err)
			time.Sleep(time.Second)
			continue
		}

		log.Info("kafka message received",
			"notification_id", payload.NotificationID,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)

		if err := svc.ProcessDelivery(ctx, payload.NotificationID); err != nil {
			if errors.Is(err, service.ErrNotFound) {
				log.Warn("notification missing; committing offset to skip poison message",
					"notification_id", payload.NotificationID,
				)
			} else {
				log.Error("process delivery failed; will not commit offset",
					"notification_id", payload.NotificationID,
					"error", err,
				)
				time.Sleep(time.Second)
				continue
			}
		} else {
			log.Info("notification marked SENT (email stub)",
				"notification_id", payload.NotificationID,
			)
		}

		if err := consumer.Commit(ctx, msg); err != nil {
			log.Error("kafka commit failed", "error", err, "offset", msg.Offset)
		}
	}
}
