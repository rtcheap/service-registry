package main

import (
	"fmt"
	"net/http"

	"github.com/CzarSimon/httputil"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	tracelog "github.com/opentracing/opentracing-go/log"
	"github.com/rtcheap/dto"
)

func (e *env) registerService(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller.registerService")
	defer span.Finish()

	var body dto.Service
	err := c.BindJSON(&body)
	if err != nil {
		err = httputil.BadRequestError(fmt.Errorf("failed to parse request body. %w", err))
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		c.Error(err)
		return
	}

	svc, err := e.registry.Register(ctx, body)
	if err != nil {
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		c.Error(err)
		return
	}

	span.LogFields(tracelog.Bool("success", true))
	c.JSON(http.StatusOK, svc)
}

func (e *env) findService(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller.findService")
	defer span.Finish()

	svc, err := e.registry.Find(ctx, c.Param("id"))
	if err != nil {
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		c.Error(err)
		return
	}

	span.LogFields(tracelog.Bool("success", true))
	c.JSON(http.StatusOK, svc)
}

func (e *env) setServiceStatus(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller.setServiceStatus")
	defer span.Finish()

	status := dto.ServiceStatus(c.Param("status"))
	err := e.registry.SetStatus(ctx, c.Param("id"), status)
	if err != nil {
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		c.Error(err)
		return
	}

	span.LogFields(tracelog.Bool("success", true))
	httputil.SendOK(c)
}
