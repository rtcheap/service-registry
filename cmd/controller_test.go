package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CzarSimon/httputil/dbutil"
	"github.com/CzarSimon/httputil/jwt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rtcheap/service-registry/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegister(t *testing.T) {
	assert := assert.New(t)
	e, ctx := createTestEnv()
	repo := repository.NewServiceRepository(e.db)

	services, err := repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 0)
}

// ---- Test utils ----

func createTestEnv() (*env, context.Context) {
	cfg := config{
		db:             dbutil.SqliteConfig{},
		migrationsPath: "../resources/db/sqlite",
		jwtCredentials: getTestJWTCredentials(),
	}

	db := dbutil.MustConnect(cfg.db)

	err := dbutil.Upgrade(cfg.migrationsPath, cfg.db.Driver(), db)
	if err != nil {
		log.Panic("Failed to apply migratons", zap.Error(err))
	}

	e := &env{
		cfg: cfg,
		db:  db,
	}

	return e, context.Background()
}

func performTestRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createTestRequest(route, method string, body interface{}) *http.Request {
	var reqBody io.Reader
	if body != nil {
		bytesBody, err := json.Marshal(body)
		if err != nil {
			log.Fatal("Failed to marshal body", zap.Error(err))
		}
		reqBody = bytes.NewBuffer(bytesBody)
	}

	req, err := http.NewRequest(method, route, reqBody)
	if err != nil {
		log.Fatal("Failed to create request", zap.Error(err))
	}

	return req
}

func getTestJWTCredentials() jwt.Credentials {
	return jwt.Credentials{
		Issuer: "chatbot-test",
		Secret: "very-secret-secret",
	}
}
