package handler

import (
	"nats/internal/service"

	"github.com/labstack/echo/v4"
)

func AccountBaseHandlers(queueSvc service.QueueService) map[string]func() echo.HandlerFunc {
	queueHandler := NewQueueHandler(queueSvc)

	return map[string]func() echo.HandlerFunc{
		"createQueue": queueHandler.Create,
		"listQueues":  queueHandler.List,
	}
}

func AccountQueueBaseHandlers(queueSvc service.QueueService, messageSvc service.MessageService) map[string]func() echo.HandlerFunc {
	queueHandler := NewQueueHandler(queueSvc)
	messageHandler := NewMessageHandler(messageSvc)

	return map[string]func() echo.HandlerFunc{
		"deleteQueue":  queueHandler.Delete,
		"message":      messageHandler.Message,
		"messageAsync": messageHandler.MessageAsync,
		"messageCheck": messageHandler.CheckAckStatus,
	}
}
