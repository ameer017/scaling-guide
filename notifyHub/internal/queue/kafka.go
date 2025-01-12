package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// SendMessage is the Kafka payload for a notification delivery job.
// Keeping only the ID forces the worker to load the latest state from Postgres.
type SendMessage struct {
	NotificationID string `json:"notification_id"`
}

// DLQMessage is published when a notification permanently fails or is poison.
type DLQMessage struct {
	NotificationID string `json:"notification_id"`
	Reason         string `json:"reason"`
	Attempt        int    `json:"attempt"`
	FailedAt       string `json:"failed_at"`
}

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(brokersCSV, topic string) *Producer {
	brokers := splitBrokers(brokersCSV)
	return &Producer{
		topic: topic,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
}

func (p *Producer) Topic() string {
	return p.topic
}

func (p *Producer) Publish(ctx context.Context, notificationID string) error {
	payload, err := json.Marshal(SendMessage{NotificationID: notificationID})
	if err != nil {
		return fmt.Errorf("marshal kafka message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(notificationID),
		Value: payload,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("publish to kafka topic %s: %w", p.topic, err)
	}
	return nil
}

func (p *Producer) PublishDLQ(ctx context.Context, msg DLQMessage) error {
	if msg.FailedAt == "" {
		msg.FailedAt = time.Now().UTC().Format(time.RFC3339)
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal dlq message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.NotificationID),
		Value: payload,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("publish to dlq topic %s: %w", p.topic, err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokersCSV, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        splitBrokers(brokersCSV),
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0, // manual commit after successful handling
			StartOffset:    kafka.FirstOffset,
		}),
	}
}

// Fetch reads the next message. Caller must Commit after successful processing.
func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, SendMessage, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, SendMessage{}, err
	}

	var payload SendMessage
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return msg, SendMessage{}, fmt.Errorf("unmarshal kafka message: %w", err)
	}
	if payload.NotificationID == "" {
		return msg, SendMessage{}, fmt.Errorf("empty notification_id in kafka message")
	}
	return msg, payload, nil
}

func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func splitBrokers(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
