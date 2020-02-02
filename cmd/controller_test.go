package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/client/rpc"
	"github.com/CzarSimon/httputil/dbutil"
	"github.com/CzarSimon/httputil/id"
	"github.com/CzarSimon/httputil/jwt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/opentracing/opentracing-go"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/service-registry/internal/repository"
	"github.com/rtcheap/service-registry/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegister_NewService(t *testing.T) {
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
	req := createTestRequest("/v1/services", http.MethodPost, jwt.SystemRole, svc)
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
	req = createTestRequest("/v1/services", http.MethodPost, jwt.SystemRole, svc)
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

func TestRegister_ExistingService(t *testing.T) {
	assert := assert.New(t)
	e, ctx := createTestEnv()
	repo := repository.NewServiceRepository(e.db)
	server := newServer(e)

	existingSvc := dto.Service{
		ID:          id.New(),
		Application: "test-app",
		Location:    "ip-1",
		Port:        8080,
		Status:      dto.StatusHealty,
	}
	_, err := repo.Save(ctx, existingSvc)
	assert.NoError(err)

	services, err := repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 1)

	svc := dto.Service{
		ID:          existingSvc.ID,
		Application: "test-app",
		Location:    "ip-2",
		Port:        8080,
		Status:      dto.StatusHealty,
	}
	req := createTestRequest("/v1/services", http.MethodPost, jwt.SystemRole, svc)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	var resBody dto.Service
	err = rpc.DecodeJSON(res.Result(), &resBody)
	assert.NoError(err)
	assert.Equal(existingSvc.ID, resBody.ID)
	assert.Equal(svc.Application, resBody.Application)
	assert.Equal(svc.Location, resBody.Location)
	assert.Equal(svc.Port, resBody.Port)
	assert.Equal(svc.Status, resBody.Status)

	services, err = repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 1)
	storedSvc := services[0]
	assert.Equal(existingSvc.ID, storedSvc.ID)
	assert.Equal(svc.Application, storedSvc.Application)
	assert.Equal(svc.Location, storedSvc.Location)
	assert.Equal(svc.Port, storedSvc.Port)
	assert.Equal(svc.Status, storedSvc.Status)

	svc = dto.Service{
		Application: "test-app",
		Location:    "ip-2",
		Port:        8080,
		Status:      dto.StatusUnhealthy,
	}
	req = createTestRequest("/v1/services", http.MethodPost, jwt.SystemRole, svc)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	err = rpc.DecodeJSON(res.Result(), &resBody)
	assert.NoError(err)
	assert.Equal(existingSvc.ID, resBody.ID)
	assert.Equal(svc.Application, resBody.Application)
	assert.Equal(svc.Location, resBody.Location)
	assert.Equal(svc.Port, resBody.Port)
	assert.Equal(svc.Status, resBody.Status)

	services, err = repo.FindByApplication(ctx, "test-app")
	assert.NoError(err)
	assert.Len(services, 1)
	storedSvc = services[0]
	assert.Equal(existingSvc.ID, storedSvc.ID)
	assert.Equal(svc.Application, storedSvc.Application)
	assert.Equal(svc.Location, storedSvc.Location)
	assert.Equal(svc.Port, storedSvc.Port)
	assert.Equal(svc.Status, storedSvc.Status)
}

func TestFindService(t *testing.T) {
	assert := assert.New(t)
	e, ctx := createTestEnv()
	repo := repository.NewServiceRepository(e.db)
	server := newServer(e)

	svcID := id.New()
	svc := dto.Service{
		ID:          svcID,
		Application: "test-app",
		Location:    "ip-1",
		Port:        8080,
		Status:      dto.StatusHealty,
	}
	_, err := repo.Save(ctx, svc)
	assert.NoError(err)
	req := createTestRequest("/v1/services/"+svcID, http.MethodGet, jwt.SystemRole, nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	var resBody dto.Service
	err = rpc.DecodeJSON(res.Result(), &resBody)
	assert.NoError(err)
	assert.Equal(svcID, resBody.ID)
	assert.Equal(svc.Application, resBody.Application)
	assert.Equal(svc.Location, resBody.Location)
	assert.Equal(svc.Port, resBody.Port)
	assert.Equal(svc.Status, resBody.Status)

	req = createTestRequest("/v1/services/"+id.New(), http.MethodGet, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusNotFound, res.Code)
}

func TestSetServiceServiceStatus(t *testing.T) {
	assert := assert.New(t)
	e, ctx := createTestEnv()
	repo := repository.NewServiceRepository(e.db)
	server := newServer(e)

	svcID := id.New()
	svc := dto.Service{
		ID:          svcID,
		Application: "test-app",
		Location:    "ip-1",
		Port:        8080,
		Status:      dto.StatusHealty,
	}
	_, err := repo.Save(ctx, svc)
	assert.NoError(err)

	var newStatus dto.ServiceStatus = dto.StatusUnhealthy
	path := fmt.Sprintf("/v1/services/%s/status/%s", svcID, newStatus)
	req := createTestRequest(path, http.MethodPut, jwt.SystemRole, nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	storedSvc, err := repo.Find(ctx, svcID)
	assert.NoError(err)
	assert.Equal(svcID, storedSvc.ID)
	assert.Equal(svc.Application, storedSvc.Application)
	assert.Equal(svc.Location, storedSvc.Location)
	assert.Equal(svc.Port, storedSvc.Port)
	assert.Equal(newStatus, storedSvc.Status)

	path = fmt.Sprintf("/v1/services/%s/status/%s", id.New(), newStatus)
	req = createTestRequest(path, http.MethodPut, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusPreconditionRequired, res.Code)

	var httpErr httputil.Error
	err = rpc.DecodeJSON(res.Result(), &httpErr)
	assert.Equal("Precondition Required", httpErr.Message)
	assert.Equal(http.StatusPreconditionRequired, httpErr.Status)
	assert.Nil(httpErr.Err)
}

func TestFindApplicationServices(t *testing.T) {
	assert := assert.New(t)
	e, ctx := createTestEnv()
	repo := repository.NewServiceRepository(e.db)
	server := newServer(e)

	storedServices := []dto.Service{
		dto.Service{
			Application: "test-app",
			Location:    "ip-1",
			Port:        8080,
			Status:      dto.StatusHealty,
		},
		dto.Service{
			Application: "other-app",
			Location:    "ip-2",
			Port:        8080,
			Status:      dto.StatusHealty,
		},
		dto.Service{
			Application: "test-app",
			Location:    "ip-3",
			Port:        8080,
			Status:      dto.StatusHealty,
		},
		dto.Service{
			Application: "test-app",
			Location:    "ip-1",
			Port:        8081,
			Status:      dto.StatusHealty,
		},
		dto.Service{
			Application: "test-app",
			Location:    "ip-4",
			Port:        8080,
			Status:      dto.StatusUnhealthy,
		},
	}
	for i, svc := range storedServices {
		svc.ID = strconv.Itoa(i + 1)
		_, err := repo.Save(ctx, svc)
		assert.NoError(err)
	}

	// Testcase: Happy path - Default to only healthy
	req := createTestRequest("/v1/services?application=test-app", http.MethodGet, jwt.SystemRole, nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	healthyServices := make([]dto.Service, 0, 3)
	err := rpc.DecodeJSON(res.Result(), &healthyServices)
	assert.NoError(err)

	assert.Len(healthyServices, 3)
	for i, expectedID := range []string{"1", "3", "4"} {
		assert.Equal(expectedID, healthyServices[i].ID)
	}

	// Testcase: Happy path - Specified only healthy
	req = createTestRequest("/v1/services?application=test-app&only-healthy=true", http.MethodGet, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	explicitHealthyServices := make([]dto.Service, 0, 3)
	err = rpc.DecodeJSON(res.Result(), &explicitHealthyServices)
	assert.NoError(err)

	assert.Len(explicitHealthyServices, 3)
	for i, expectedID := range []string{"1", "3", "4"} {
		assert.Equal(expectedID, explicitHealthyServices[i].ID)
	}

	// Testcase: Happy path - Specified to include unhealthy
	req = createTestRequest("/v1/services?application=test-app&only-healthy=0", http.MethodGet, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	servicesIncludingUnhealthy := make([]dto.Service, 0, 4)
	err = rpc.DecodeJSON(res.Result(), &servicesIncludingUnhealthy)
	assert.NoError(err)

	assert.Len(servicesIncludingUnhealthy, 4)
	for i, expectedID := range []string{"1", "3", "4", "5"} {
		assert.Equal(expectedID, servicesIncludingUnhealthy[i].ID)
	}

	// Testcase: Happy path - No services exist for application
	req = createTestRequest("/v1/services?application=missing-app", http.MethodGet, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	noServices := make([]dto.Service, 0)
	err = rpc.DecodeJSON(res.Result(), &noServices)
	assert.NoError(err)
	assert.Len(noServices, 0)

	// Testcase: No application specified, should return 400 error
	req = createTestRequest("/v1/services", http.MethodGet, jwt.SystemRole, nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusBadRequest, res.Code)
}

func TestHealthCheck(t *testing.T) {
	assert := assert.New(t)
	e, _ := createTestEnv()
	server := newServer(e)

	req := createTestRequest("/health", http.MethodGet, "", nil)
	res := performTestRequest(server.Handler, req)
	assert.Equal(http.StatusOK, res.Code)

	err := e.db.Close()
	assert.NoError(err)

	// Database disconnected, should return 503 error
	req = createTestRequest("/health", http.MethodGet, "", nil)
	res = performTestRequest(server.Handler, req)
	assert.Equal(http.StatusServiceUnavailable, res.Code)
}

func TestPermissions(t *testing.T) {
	assert := assert.New(t)
	e, _ := createTestEnv()
	server := newServer(e)

	cases := []struct {
		method string
		route  string
	}{
		{method: http.MethodPost, route: "/v1/services"},
		{method: http.MethodGet, route: "/v1/services/some-id"},
		{method: http.MethodGet, route: "/v1/services?application=some-app"},
		{method: http.MethodPut, route: "/v1/services/some-id/status/HEALTHY"},
	}

	badRoles := []string{jwt.AnonymousRole, jwt.AdminRole, ""}

	for _, tc := range cases {
		for _, role := range badRoles {
			req := createTestRequest(tc.route, tc.method, role, nil)
			res := performTestRequest(server.Handler, req)

			expectedStatus := http.StatusForbidden
			if role == "" {
				expectedStatus = http.StatusUnauthorized
			}
			assert.Equal(expectedStatus, res.Code)
		}
	}
}

// ---- Test utils ----

func createTestEnv() (*env, context.Context) {
	cfg := config{
		db:             dbutil.SqliteConfig{},
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
