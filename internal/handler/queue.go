package handler

import (
	"nats/internal/context/logs"
	"nats/internal/entity"
	"nats/internal/service"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type QueueHandler struct {
	svc service.QueueService
}

func NewQueueHandler(svc service.QueueService) *QueueHandler {
	return &QueueHandler{svc: svc}
}

type CreateQueueRequest struct {
	Name string `json:"Name" validate:"required"`
}

type CreateQueueResponse struct {
	CreateQueueResult entity.Queue            `json:"CreateQueueResult"`
	ResponseMetadata  entity.ResponseMetadata `json:"ResponseMetadata"`
}

type DeleteQueueRequest struct {
	QueueSrn string `json:"QueueSrn" validate:"required"`
}

type DeleteQueueResponse struct {
	ResponseMetadata entity.ResponseMetadata `json:"ResponseMetadata"`
}

type ListQueuesResponse struct {
	Queues []entity.Queue `json:"queues"`
}

func (h *QueueHandler) Create() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var req CreateQueueRequest
		if err := c.Bind(&req); err != nil {
			logs.GetLogger(ctx).Error("Invalid createQueue request parameter", zap.Error(err))
			return c.JSON(entity.InvalidParameter.HTTPCode, entity.InvalidParameter.Error)
		}

		if err := c.Validate(&req); err != nil {
			logs.GetLogger(ctx).Error("Required parameter is missing", zap.Error(err))
			return c.JSON(entity.InvalidParameter.HTTPCode, entity.InvalidParameter.Error)
		}

		result, err := h.svc.CreateQueue(ctx, req.Name, c.Param("accountid"))
		if err != nil {
			logs.GetLogger(ctx).Error("Failed to create stream", zap.Error(err))
			return c.JSON(entity.InternalError.HTTPCode, entity.InternalError.Error)
		}
		logs.GetLogger(ctx).Info("Stream creation success", zap.String("queue", req.Name))
		meta := entity.ResponseMetadata{RequestId: c.Response().Header().Get(echo.HeaderXRequestID)}
		return c.JSON(http.StatusOK, CreateQueueResponse{
			CreateQueueResult: result, ResponseMetadata: meta,
		})
	}
}

func (h *QueueHandler) Delete() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var req DeleteQueueRequest
		if err := c.Bind(&req); err != nil {
			logs.GetLogger(ctx).Error("Invalid deleteQueue request parameter", zap.Error(err))
			return c.JSON(entity.InvalidParameter.HTTPCode, entity.InvalidParameter.Error)
		}

		if err := c.Validate(&req); err != nil {
			logs.GetLogger(ctx).Error("Required parameter is missing", zap.Error(err))
			return c.JSON(entity.InvalidParameter.HTTPCode, entity.InvalidParameter.Error)
		}

		parts := strings.Split(req.QueueSrn, ":")
		name := parts[len(parts)-1]

		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing 'name' parameter"})
		}

		if err := h.svc.DeleteQueue(ctx, name); err != nil {
			logs.GetLogger(ctx).Error("Failed to delete stream", zap.Error(err))
			return c.JSON(entity.InternalError.HTTPCode, entity.InternalError.Error)
		}

		logs.GetLogger(ctx).Info("Stream deletion success", zap.String("queue", name))
		meta := entity.ResponseMetadata{RequestId: c.Response().Header().Get(echo.HeaderXRequestID)}
		return c.JSON(http.StatusOK, DeleteQueueResponse{ResponseMetadata: meta})
	}
}

func (h *QueueHandler) List() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		queues, err := h.svc.ListQueues(ctx, c.Param("accountid"))
		if err != nil {
			logs.GetLogger(ctx).Error("Queue list lookup failed", zap.Error(err))
			return c.JSON(entity.InternalError.HTTPCode, entity.InternalError.Error)
		}

		logs.GetLogger(ctx).Info("Return queue list", zap.Int("count", len(queues)))
		return c.JSON(http.StatusOK, ListQueuesResponse{Queues: queues})
	}
}
