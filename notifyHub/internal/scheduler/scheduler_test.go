package scheduler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type fakeClaimer struct {
	ids []string
	err error
}

func (f *fakeClaimer) ClaimDueScheduled(ctx context.Context, limit int) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.ids) > limit {
		return f.ids[:limit], nil
	}
	return f.ids, nil
}

type fakePublisher struct {
	published []string
	err       error
}

func (f *fakePublisher) Publish(ctx context.Context, notificationID string) error {
	if f.err != nil {
		return f.err
	}
	f.published = append(f.published, notificationID)
	return nil
}

func TestSchedulerTickPublishesClaimedIDs(t *testing.T) {
	t.Parallel()

	claimer := &fakeClaimer{ids: []string{"a", "b"}}
	publisher := &fakePublisher{}
	s := New(claimer, publisher, time.Second, 10, slog.New(slog.NewTextHandler(io.Discard, nil)))

	s.tick(context.Background())

	if len(publisher.published) != 2 {
		t.Fatalf("published = %v, want 2 ids", publisher.published)
	}
}

func TestSchedulerTickSkipsWhenClaimEmpty(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{}
	s := New(&fakeClaimer{}, publisher, time.Second, 10, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.tick(context.Background())
	if len(publisher.published) != 0 {
		t.Fatalf("expected no publishes, got %v", publisher.published)
	}
}

func TestSchedulerTickHandlesClaimError(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{}
	s := New(
		&fakeClaimer{err: errors.New("db down")},
		publisher,
		time.Second,
		10,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	s.tick(context.Background())
	if len(publisher.published) != 0 {
		t.Fatalf("expected no publishes on claim error")
	}
}

func TestSchedulerTickContinuesAfterPublishError(t *testing.T) {
	t.Parallel()

	claimer := &fakeClaimer{ids: []string{"a", "b"}}
	publisher := &fakePublisher{err: errors.New("kafka down")}
	s := New(claimer, publisher, time.Second, 10, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.tick(context.Background())
	// first publish fails; loop continues but none succeed
	if len(publisher.published) != 0 {
		t.Fatalf("expected no successful publishes")
	}
}
