package rpc

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/segmentio/kafka-go"
	"internal/advert"
	"internal/env"
	"internal/global"
)

type MbHandler struct {
	hub global.Hub
}

func NewMbHandler(hub global.Hub) *MbHandler {
	return &MbHandler{hub: hub}
}

func (h *MbHandler) Handle(logger logr.Logger, m *kafka.Message) error {
	env := env.NewEnvironment(h.hub)
	env.Logger = logger.WithName(fmt.Sprintf("[message broker][message][%s][%s]", m.Topic, m.Key))
	defer env.Close()

	switch m.Topic {
	case "advert_process_photo_response":
		return advert.ResponsePhotoProcess(env, m)
	}
	return nil
}
