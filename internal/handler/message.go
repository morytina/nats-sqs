package handler

import (
	"nats/internal/context/logs"
	"nats/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type MessageHandler struct {
	svc service.MessageService
}

func NewMessageHandler(svc service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

type MessageRequest struct {
	QueueName string `json:"queueName"`
	Message   string `json:"message"`
	Subject   string `json:"subject"`
}

type MessageResponse struct {
	Sequence uint64 `json:"sequence"`
}

type MessageAsyncResponse struct {
	MessageID string `json:"messageId"`
}

func (h *MessageHandler) Message() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logs.GetLogger(ctx)

		var req MessageRequest
		if err := c.Bind(&req); err != nil {
			logger.Warn("메시지 요청 파싱 실패", zap.Error(err))
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		seq, err := h.svc.SendMessage(ctx, req.QueueName, req.Message, req.Subject)
		if err != nil {
			logger.Error("메시지 발행 실패", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		logger.Info("메시지 발행 성공", zap.Uint64("sequence", seq))
		return c.JSON(http.StatusOK, MessageResponse{Sequence: seq})
	}
}

func (h *MessageHandler) MessageAsync() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logs.GetLogger(ctx)

		var req MessageRequest
		if err := c.Bind(&req); err != nil {
			logger.Warn("메시지 요청 파싱 실패", zap.Error(err))
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		msgID, err := h.svc.SendAsyncMessage(ctx, req.QueueName, req.Message, req.Subject)
		if err != nil {
			logger.Error("메시지 발행 실패", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		logger.Info("메시지 발행 성공", zap.String("messageId", msgID))
		return c.JSON(http.StatusOK, MessageAsyncResponse{MessageID: msgID})
	}
}

func (h *MessageHandler) CheckAckStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logs.GetLogger(ctx)

		id := c.QueryParam("messageId")
		if id == "" {
			logger.Warn("ack 조회 요청에 ID 없음")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing message id"})
		}

		status, err := h.svc.CheckAckStatus(ctx, id)
		if err != nil {
			logger.Warn("ack 상태 조회 실패", zap.String("id", id), zap.Error(err))
			return c.JSON(http.StatusNotFound, map[string]string{"error": "message id not found"})
		}

		logger.Info("ack 상태 조회 성공", zap.String("id", id), zap.String("status", status))
		return c.JSON(http.StatusOK, map[string]string{"status": status})
	}
}
