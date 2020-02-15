package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CzarSimon/httputil"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	tracelog "github.com/opentracing/opentracing-go/log"
	"github.com/rtcheap/dto"
)

func (e *env) registerService(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller_register_service")
	defer span.Finish()

	var body dto.Service
	err := c.BindJSON(&body)
	if err != nil {
		err = httputil.BadRequestError(fmt.Errorf("failed to parse request body. %w", err))
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	svc, err := e.registry.Register(ctx, body)
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, svc)
}

func (e *env) findService(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller_find_service")
	defer span.Finish()

	svc, err := e.registry.Find(ctx, c.Param("id"))
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, svc)
}

func (e *env) setServiceStatus(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller_set_service_status")
	defer span.Finish()

	status := dto.ServiceStatus(c.Param("status"))
	err := e.registry.SetStatus(ctx, c.Param("id"), status)
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	httputil.SendOK(c)
}

func (e *env) findApplicationServices(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "controller_find_application_services")
	defer span.Finish()

	onlyHealthy := parseQueryFlag(c, "only-healthy", true)
	application, err := httputil.ParseQueryValue(c, "application")
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	services, err := e.registry.FindApplicationServices(ctx, application, onlyHealthy)
	if err != nil {
		span.LogFields(tracelog.Error(err))
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, services)
}

func parseQueryFlag(c *gin.Context, name string, defaultValue bool) bool {
	flag, ok := c.GetQuery(name)
	if !ok {
		return defaultValue
	}

	return strings.ToLower(flag) == "true" || flag == "1"
}
