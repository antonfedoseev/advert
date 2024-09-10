package mb

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	ctx            context.Context
	brokers        []string
	settings       ConsumerSpec
	logger         logr.Logger
	messageHandler MessageHandler
	reader         *kafka.Reader
	stop           chan struct{}
	messages       chan *kafka.Message
}

type MessageHandler interface {
	Handle(logger logr.Logger, m *kafka.Message) error
}

func NewConsumer(ctx context.Context, settings Settings, logger logr.Logger, messageHandler MessageHandler) *Consumer {
	c := &Consumer{
		ctx:            ctx,
		settings:       settings.Consumer,
		brokers:        settings.Brokers,
		logger:         logger.WithName("[message broker][consumer]"),
		messageHandler: messageHandler,
		stop:           make(chan struct{}),
		messages:       make(chan *kafka.Message, settings.Consumer.WorkersAmount),
	}
	c.setupReader()
	go c.createWorkers()
	go c.read()

	return c
}

func (c *Consumer) setupReader() {
	config := kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     c.settings.GroupId,
		MaxAttempts: c.settings.ReadRetries,
		GroupTopics: c.settings.Topics,
	}
	c.reader = kafka.NewReader(config)
}

func (c *Consumer) read() {
	defer close(c.messages)
	for {
		select {
		case <-c.stop:
			return
		case <-c.ctx.Done():
			return
		default:
			m, err := c.reader.FetchMessage(c.ctx)
			if err != nil {
				return
			}
			c.messages <- &m
		}
	}
}

func (c *Consumer) Close() {
	close(c.stop)
}

func (c *Consumer) createWorkers() {
	for i := 0; i < c.settings.WorkersAmount; i++ {
		go c.handle()
	}
}

func (c *Consumer) handle() {
	for {
		select {
		case <-c.stop:
			return
		case <-c.ctx.Done():
			return
		case m, ok := <-c.messages:
			if !ok {
				return
			}
			c.handleMessage(m)
		}
	}
}

func (c *Consumer) handleMessage(m *kafka.Message) {
	err := c.messageHandler.Handle(c.logger, m)
	if err != nil {
		c.logger.Error(err, fmt.Sprintf("error of handling message: \"%v\".", *m))
		return
	}

	if err := c.reader.CommitMessages(c.ctx, *m); err != nil {
		c.logger.Error(err, fmt.Sprintf("failed to commit message: \"%v\".", *m))
	}
}
