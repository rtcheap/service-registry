package main

import (
	"fmt"
	"net/http"

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
		err = fmt.Errorf("failed to parse request body. %w", err)
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

	span.LogFields(tracelog.Bool("success", false))
	c.JSON(http.StatusOK, svc)
}
