package main

import (
	"net/http"
	"time"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/jwt"
	"github.com/CzarSimon/httputil/logger"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var log = logger.GetDefaultLogger("service-registry/main")

func main() {
	e := setupEnv()
	defer e.close()

	server := newServer(e)
	log.Info("Started service-registry listening on port: " + e.cfg.port)

	err := server.ListenAndServe()
	if err != nil {
		log.Error("Unexpected error stoped server.", zap.Error(err))
	}
}

func newServer(e *env) *http.Server {
	r := httputil.NewRouter("service-registry", e.checkHealth)

	rbac := httputil.RBAC{
		Verifier: jwt.NewVerifier(e.cfg.jwtCredentials, time.Minute),
	}
	v1 := r.Group("/v1", rbac.Secure(jwt.SystemRole))

	v1.POST("/services", e.registerService)
	v1.GET("/services", e.findApplicationServices)
	v1.GET("/services/:id", e.findService)
	v1.PUT("/services/:id/status/:status", e.setServiceStatus)

	return &http.Server{
		Addr:    ":" + e.cfg.port,
		Handler: r,
	}
}
