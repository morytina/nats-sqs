package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"nats/internal/context/logs"
	"nats/internal/context/metrics"
)

type ApiRouter interface {
	Register(g *echo.Group)
}

type apiRouter struct {
	accountBaseHandlers      map[string]func() echo.HandlerFunc
	accountQueueBaseHandlers map[string]func() echo.HandlerFunc
}

func NewApiRouter(accountBaseHandlers map[string]func() echo.HandlerFunc, accountQueueBaseHandlers map[string]func() echo.HandlerFunc) ApiRouter {
	return &apiRouter{accountBaseHandlers: accountBaseHandlers, accountQueueBaseHandlers: accountQueueBaseHandlers}
}

func (r *apiRouter) Register(g *echo.Group) {
	g.Any("/:accountid", r.handleAccountBase)
	g.Any("/:accountid/:queueid", r.handleAccountQueueBase)
}

func (r *apiRouter) handleAccountBase(c echo.Context) error {
	logs.GetLogger(c.Request().Context()).Info("handleAccountBase")
	action := c.QueryParam("Action")

	if handlerFunc, ok := r.accountBaseHandlers[action]; ok {
		err := handlerFunc()(c)
		metrics.ApiCallCounter.WithLabelValues(action, strconv.Itoa(c.Response().Status)).Inc()
		return err
	}

	metrics.ApiCallCounter.WithLabelValues(action, "400").Inc()
	return c.String(http.StatusBadRequest, "invalid Action")
}

func (r *apiRouter) handleAccountQueueBase(c echo.Context) error {
	logs.GetLogger(c.Request().Context()).Info("handleAccountQueueBase")
	action := c.QueryParam("Action")

	if handlerFunc, ok := r.accountQueueBaseHandlers[action]; ok {
		err := handlerFunc()(c)
		metrics.ApiCallCounter.WithLabelValues(action, strconv.Itoa(c.Response().Status)).Inc()
		return err
	}

	metrics.ApiCallCounter.WithLabelValues(action, "400").Inc()
	return c.String(http.StatusBadRequest, "invalid Action")
}
