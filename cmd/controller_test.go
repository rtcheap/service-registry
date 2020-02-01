package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CzarSimon/httputil/client/rpc"
	"github.com/CzarSimon/httputil/dbutil"
	"github.com/CzarSimon/httputil/jwt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/opentracing/opentracing-go"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/service-registry/internal/repository"
	"github.com/rtcheap/service-registry/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegister(t *testing.T) {
	assert := assert.New(t)
	e, ctx := createTestEnv()
	repo := repository.NewServiceRepository(e.db)
	server := newServer(e)

	services, err := repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 0)

	svc := dto.Service{
		Application: "test-app",
		Location:    "ip-1",
		Port:        8080,
		Status:      dto.StatusHealty,
	}
	req := createTestRequest("/v1/services", http.MethodPost, jwt.AnonymousRole, svc)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	var resBody dto.Service
	err = rpc.DecodeJSON(res.Result(), &resBody)
	assert.NoError(err)
	assert.NotEqual("", resBody.ID)
	assert.Equal(svc.Application, resBody.Application)
	assert.Equal(svc.Location, resBody.Location)
	assert.Equal(svc.Port, resBody.Port)
	assert.Equal(svc.Status, resBody.Status)

	services, err = repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 1)
	storedSvc := services[0]
	assert.Equal(resBody.ID, storedSvc.ID)
	assert.Equal(svc.Application, storedSvc.Application)
	assert.Equal(svc.Location, storedSvc.Location)
	assert.Equal(svc.Port, storedSvc.Port)
	assert.Equal(svc.Status, storedSvc.Status)

	svc = dto.Service{
		Application: "test-app",
		Location:    "ip-2",
		Port:        8080,
	}
	req = createTestRequest("/v1/services", http.MethodPost, jwt.AnonymousRole, svc)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	err = rpc.DecodeJSON(res.Result(), &resBody)
	assert.NoError(err)
	assert.NotEqual("", resBody.ID)
	assert.Equal(svc.Application, resBody.Application)
	assert.Equal(svc.Location, resBody.Location)
	assert.Equal(svc.Port, resBody.Port)
	assert.Equal(dto.StatusHealty, resBody.Status)

	services, err = repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 2)
	storedSvc = services[1]
	assert.Equal(resBody.ID, storedSvc.ID)
	assert.Equal(svc.Application, storedSvc.Application)
	assert.Equal(svc.Location, storedSvc.Location)
	assert.Equal(svc.Port, storedSvc.Port)
	assert.Equal(dto.StatusHealty, storedSvc.Status)
}

// ---- Test utils ----

func createTestEnv() (*env, context.Context) {
	cfg := config{
		db: dbutil.SqliteConfig{
			Name: "./test.db",
		},
		migrationsPath: "../resources/db/sqlite",
		jwtCredentials: getTestJWTCredentials(),
	}

	db := dbutil.MustConnect(cfg.db)

	err := dbutil.Downgrade(cfg.migrationsPath, cfg.db.Driver(), db)
	if err != nil {
		log.Panic("Failed to apply downgrade migratons", zap.Error(err))
	}

	err = dbutil.Upgrade(cfg.migrationsPath, cfg.db.Driver(), db)
	if err != nil {
		log.Panic("Failed to apply upgrade migratons", zap.Error(err))
	}

	repo := repository.NewServiceRepository(db)

	e := &env{
		cfg:      cfg,
		db:       db,
		registry: service.NewRegistryService(repo),
	}

	return e, context.Background()
}

func performTestRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createTestRequest(route, method, role string, body interface{}) *http.Request {
	client := rpc.NewClient(time.Second)
	req, err := client.CreateRequest(method, route, body)
	if err != nil {
		log.Fatal("Failed to create request", zap.Error(err))
	}

	span := opentracing.StartSpan(fmt.Sprintf("%s.%s", method, route))
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	if role == "" {
		return req
	}

	issuer := jwt.NewIssuer(getTestJWTCredentials())
	token, err := issuer.Issue(jwt.User{
		ID:    "service-registry-user",
		Roles: []string{role},
	}, time.Hour)

	req.Header.Add("Authorization", "Bearer "+token)
	return req
}

func getTestJWTCredentials() jwt.Credentials {
	return jwt.Credentials{
		Issuer: "service-registry-test",
		Secret: "very-secret-secret",
	}
}
