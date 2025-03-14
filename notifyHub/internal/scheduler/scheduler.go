package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Publisher publishes a notification ID to Kafka.
type Publisher interface {
	Publish(ctx context.Context, notificationID string) error
}

// Claimer claims due scheduled notifications and returns their IDs.
type Claimer interface {
	ClaimDueScheduled(ctx context.Context, limit int) ([]string, error)
}

// Scheduler polls for due SCHEDULED notifications and publishes them.
type Scheduler struct {
	claimer   Claimer
	publisher Publisher
	interval  time.Duration
	batchSize int
	log       *slog.Logger
}

func New(claimer Claimer, publisher Publisher, interval time.Duration, batchSize int, log *slog.Logger) *Scheduler {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 50
	}
	return &Scheduler{
		claimer:   claimer,
		publisher: publisher,
		interval:  interval,
		batchSize: batchSize,
		log:       log,
	}
}

// Run blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("scheduler starting", "interval", s.interval.String(), "batch_size", s.batchSize)

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	ids, err := s.claimer.ClaimDueScheduled(ctx, s.batchSize)
	if err != nil {
		s.log.Error("scheduler claim failed", "error", err)
		return
	}
	if len(ids) == 0 {
		return
	}

	s.log.Info("scheduler claimed due notifications", "count", len(ids))
	for _, id := range ids {
		if err := s.publisher.Publish(ctx, id); err != nil {
			s.log.Error("scheduler publish failed", "notification_id", id, "error", err)
			continue
		}
		s.log.Info("scheduler published due notification", "notification_id", id)
	}
}
