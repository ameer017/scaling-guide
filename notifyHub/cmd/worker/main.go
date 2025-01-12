package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notifyHub/internal/config"
	"notifyHub/internal/email"
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

	mailer, err := email.NewSMTPSender(email.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
	})
	if err != nil {
		log.Error("smtp mailer config invalid", "error", err)
		os.Exit(1)
	}

	producer := queue.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error("kafka producer close failed", "error", err)
		}
	}()

	dlq := queue.NewProducer(cfg.KafkaBrokers, cfg.KafkaDLQTopic)
	defer func() {
		if err := dlq.Close(); err != nil {
			log.Error("kafka dlq producer close failed", "error", err)
		}
	}()

	svc := service.NewNotificationService(service.Deps{
		Repo:        repository.NewNotificationRepository(pool),
		Logs:        repository.NewDeliveryLogRepository(pool),
		Producer:    producer,
		DLQ:         dlq,
		Mailer:      mailer,
		MaxAttempts: cfg.MaxDeliveryAttempts,
		BackoffBase: cfg.RetryBackoff,
		Provider:    "gmail-smtp",
	})

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
		"kafka_dlq_topic", cfg.KafkaDLQTopic,
		"kafka_group_id", cfg.KafkaGroupID,
		"max_attempts", cfg.MaxDeliveryAttempts,
		"retry_backoff", cfg.RetryBackoff.String(),
		"smtp_host", cfg.SMTPHost,
		"smtp_from", firstNonEmpty(cfg.SMTPFrom, cfg.SMTPUsername),
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

		outcome, err := svc.ProcessDelivery(ctx, payload.NotificationID)
		if err != nil && !errors.Is(err, service.ErrNotFound) &&
			!errors.Is(err, service.ErrDeliveryFailed) &&
			!errors.Is(err, service.ErrRetryScheduled) {
			log.Error("process delivery failed; will not commit offset",
				"notification_id", payload.NotificationID,
				"error", err,
			)
			time.Sleep(time.Second)
			continue
		}

		switch outcome {
		case service.OutcomeSent:
			log.Info("notification emailed and marked SENT", "notification_id", payload.NotificationID)
		case service.OutcomeRetryScheduled:
			log.Warn("delivery failed; retry scheduled",
				"notification_id", payload.NotificationID,
				"error", err,
			)
		case service.OutcomeFailedPermanent:
			log.Error("delivery permanently failed; published to DLQ",
				"notification_id", payload.NotificationID,
				"error", err,
			)
		case service.OutcomeSkipped:
			log.Info("notification already terminal; skipping", "notification_id", payload.NotificationID)
		}

		if err := consumer.Commit(ctx, msg); err != nil {
			log.Error("kafka commit failed", "error", err, "offset", msg.Offset)
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
