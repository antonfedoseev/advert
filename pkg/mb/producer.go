package mb

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"net"
	"time"
)

type Producer struct {
	brokers  []string
	settings ProducerSpec
}

func NewProducer(settings Settings) *Producer {
	pool := &Producer{settings: settings.Producer, brokers: settings.Brokers}
	return pool
}

func (p *Producer) SendMessage(ctx context.Context, topic string, key string, value interface{}) error {
	writer := p.createWriter()
	defer writer.Close()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	return writeMessage(ctx, writer, time.Duration(p.settings.ConnMaxLifetimeSec)*time.Second, msg)
}

func (p *Producer) SendMessages(ctx context.Context, topics []string, key string, value interface{}) error {
	writer := p.createWriter()
	defer writer.Close()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msgs := make([]kafka.Message, 0, len(topics))
	for i := 0; i < len(topics); i++ {
		msg := kafka.Message{
			Topic: topics[i],
			Key:   []byte(key),
			Value: data,
		}
		msgs = append(msgs, msg)
	}

	return writeMessage(ctx, writer, time.Duration(p.settings.ConnMaxLifetimeSec)*time.Second, msgs...)
}

func (p *Producer) createWriter() *kafka.Writer {
	w := kafka.Writer{
		Addr:     kafka.TCP(p.brokers...),
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			Dial:        (&net.Dialer{}).DialContext,
			IdleTimeout: time.Duration(p.settings.ConnMaxIdleTimeSec) * time.Second,
			DialTimeout: time.Duration(p.settings.ConnMaxLifetimeSec) * time.Second,
		},
		MaxAttempts: p.settings.SendRetries,
	}

	return &w
}

func writeMessage(ctx context.Context, writer *kafka.Writer, timeout time.Duration, msgs ...kafka.Message) error {
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return writer.WriteMessages(ctx, msgs...)
	}

	return writer.WriteMessages(ctx, msgs...)
}
