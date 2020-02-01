package main

import (
	"database/sql"
	"net/http"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/dbutil"
	"github.com/CzarSimon/httputil/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rtcheap/service-registry/internal/service"
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

	r.POST("/v1/services", notImplemented)
	r.GET("/v1/services", notImplemented)
	r.GET("/v1/services/:id", notImplemented)
	r.PUT("/v1/services/:id/status/:status", notImplemented)

	return &http.Server{
		Addr:    ":" + e.cfg.port,
		Handler: r,
	}
}

type env struct {
	cfg      config
	db       *sql.DB
	registry *service.RegistryService
}

func (e *env) checkHealth() error {
	return dbutil.Connected(e.db)
}

func (e *env) close() {
	err := e.db.Close()
	if err != nil {
		log.Error("failed to close database connection", zap.Error(err))
	}
}

func setupEnv() *env {
	cfg := getConfig()
	db := dbutil.MustConnect(cfg.db)

	err := dbutil.Upgrade(cfg.migrationsPath, cfg.db.Driver(), db)
	if err != nil {
		log.Fatal("failed to apply database migrations", zap.Error(err))
	}

	return &env{
		cfg:      cfg,
		db:       db,
		registry: service.NewRegistryService(nil),
	}
}

func notImplemented(c *gin.Context) {
	err := httputil.NotImplementedError(nil)
	c.Error(err)
}
