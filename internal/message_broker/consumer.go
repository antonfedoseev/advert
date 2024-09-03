package message_broker

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/segmentio/kafka-go"
	"internal/env"
	"internal/global"
)

type Consumer struct {
	ctx            context.Context
	hub            global.Hub
	brokers        []string
	settings       ConsumerSpec
	logger         logr.Logger
	messageHandler MessageHandler
	reader         *kafka.Reader
	stop           chan struct{}
	messages       chan *kafka.Message
}

type MessageHandler func(env *env.Environment, m *kafka.Message) error

func NewConsumer(ctx context.Context, hub global.Hub, settings Settings, messageHandler MessageHandler) *Consumer {
	c := &Consumer{
		ctx:            ctx,
		hub:            hub,
		settings:       settings.Consumer,
		brokers:        settings.Brokers,
		logger:         hub.Logger.WithName("[message broker]"),
		messageHandler: messageHandler,
		stop:           make(chan struct{}),
		messages:       make(chan *kafka.Message, settings.Consumer.WorkersAmount),
	}
	c.setupReader()
	go c.createWorkers()
	go c.read()

	return c
}

func (c *Consumer) setupReader() *kafka.Reader {
	config := kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     c.settings.GroupId,
		MaxAttempts: c.settings.ReadRetries,
	}
	r := kafka.NewReader(config)

	return r
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
	env := env.NewEnvironment(c.hub)
	env.Logger = env.Logger.WithName(fmt.Sprintf("[message broker][message][%s][%s]", m.Topic, m.Key))
	defer env.Close()

	err := c.messageHandler(env, m)
	if err != nil {
		c.logger.Error(err, fmt.Sprintf("error of handling message: \"%v\".", *m))
		return
	}

	if err := c.reader.CommitMessages(c.ctx, *m); err != nil {
		c.logger.Error(err, fmt.Sprintf("failed to commit message: \"%v\".", *m))
	}
}
