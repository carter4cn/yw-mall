package kafka

import (
	"context"
	"time"

	kgo "github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kgo.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		w: &kgo.Writer{
			Addr:         kgo.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kgo.Hash{},
			BatchTimeout: 50 * time.Millisecond,
			RequiredAcks: kgo.RequireAll,
			Async:        false,
		},
	}
}

func (p *Producer) Write(ctx context.Context, key string, value []byte) error {
	return p.w.WriteMessages(ctx, kgo.Message{Key: []byte(key), Value: value})
}

func (p *Producer) Close() error { return p.w.Close() }
