package handlers

import (
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
)

type SubscribeEventHandler struct {
}

func NewSubscribeEventHandler() *SubscribeEventHandler {
	return &SubscribeEventHandler{}
}

func (h *SubscribeEventHandler) Handle(event messaging.SubscribeEvent) error {
	fmt.Printf("Subscribe event: %v\n", event)
	return nil
}
