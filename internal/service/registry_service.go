package service

import (
	"context"
	"fmt"

	"github.com/CzarSimon/httputil"
	"github.com/CzarSimon/httputil/id"
	"github.com/opentracing/opentracing-go"
	tracelog "github.com/opentracing/opentracing-go/log"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/service-registry/internal/repository"
)

// RegistryService service registry.
type RegistryService struct {
	repo repository.ServiceRepository
}

// NewRegistryService sets up and creates a new service repository.
func NewRegistryService(repo repository.ServiceRepository) *RegistryService {
	return &RegistryService{
		repo: repo,
	}
}

// Register saves information about a service.
func (s *RegistryService) Register(ctx context.Context, svc dto.Service) (dto.Service, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RegistryService.Register")
	defer span.Finish()

	if svc.ID == "" {
		svc.ID = id.New()
	}
	if svc.Status == "" {
		svc.Status = dto.StatusHealty
	}

	saved, err := s.repo.Save(ctx, svc)
	if err != nil {
		err = httputil.InternalServerError(err)
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		return dto.Service{}, err
	}

	span.LogFields(tracelog.Bool("success", true))
	return saved, nil
}

// SetStatus records the status of a given service.
func (s *RegistryService) SetStatus(ctx context.Context, id string, status dto.ServiceStatus) error {
	return fmt.Errorf("not implemented")
}

// FindApplicationServices looks up all serices for an application.
func (s *RegistryService) FindApplicationServices(ctx context.Context, application string, includeUnhealthy bool) ([]dto.Service, error) {
	return nil, fmt.Errorf("not implemented")
}
