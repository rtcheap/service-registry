package service

import (
	"context"
	"database/sql"
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

// Find looks up and and returns service with the given id.
func (s *RegistryService) Find(ctx context.Context, id string) (dto.Service, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RegistryService.Find")
	defer span.Finish()

	svc, err := s.repo.Find(ctx, id)
	if err != nil {
		notFound := err == sql.ErrNoRows
		span.LogFields(tracelog.Bool("success", notFound), tracelog.Error(err))
		if notFound {
			err = httputil.NotFoundError(err)
		}
		return dto.Service{}, err
	}

	span.LogFields(tracelog.Bool("success", true))
	return svc, nil
}

// SetStatus records the status of a given service.
func (s *RegistryService) SetStatus(ctx context.Context, id string, status dto.ServiceStatus) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RegistryService.SetStatus")
	defer span.Finish()

	svc, err := s.repo.Find(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = httputil.PreconditionRequiredError(err)
		}
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		return err
	}

	svc.Status = status
	_, err = s.repo.Save(ctx, svc)
	if err != nil {
		err := fmt.Errorf("failed to save status update for service(id=%s). %w", id, err)
		span.LogFields(tracelog.Bool("success", false), tracelog.Error(err))
		return err
	}

	span.LogFields(tracelog.Bool("success", true))
	return nil
}

// FindApplicationServices looks up all serices for an application.
func (s *RegistryService) FindApplicationServices(ctx context.Context, application string, includeUnhealthy bool) ([]dto.Service, error) {
	return nil, fmt.Errorf("not implemented")
}
