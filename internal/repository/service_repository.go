package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rtcheap/dto"
)

// ServiceRepository storage interface for service metadata.
type ServiceRepository interface {
	Save(ctx context.Context, svc dto.Service) error
	Find(ctx context.Context, id string) (dto.Service, error)
	FindByApplication(ctx context.Context, application string) ([]dto.Service, error)
}

// NewServiceRepository creates a service repository using the default implementation.
func NewServiceRepository(db *sql.DB) ServiceRepository {
	return &serviceRepo{
		db: db,
	}
}

type serviceRepo struct {
	db *sql.DB
}

func (r *serviceRepo) Save(ctx context.Context, svc dto.Service) error {
	return fmt.Errorf("not implemented")
}

func (r *serviceRepo) Find(ctx context.Context, id string) (dto.Service, error) {
	return dto.Service{}, fmt.Errorf("not implemented")
}

const findByApplicationQuery = `
	SELECT 
		id, 
		application, 
		location, 
		port, 
		status 
	FROM service
	WHERE 
		application = ?`

func (r *serviceRepo) FindByApplication(ctx context.Context, application string) ([]dto.Service, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serviceRepo.FindByApplication")
	defer span.Finish()

	rows, err := r.db.QueryContext(ctx, findByApplicationQuery, application)
	if err != nil {
		err = fmt.Errorf("failed to query database. %w", err)
		recordError(span, err)
		return nil, err
	}

	services := make([]dto.Service, 0)
	for rows.Next() {
		s := dto.Service{}
		err = rows.Scan(&s.ID, &s.Application, &s.Location, &s.Port, &s.Status)
		if err != nil {
			err = fmt.Errorf("failed to scan row. %w", err)
			recordError(span, err)
			return nil, err
		}

		services = append(services, s)
	}

	span.LogFields(log.Bool("success", true))
	return services, nil
}

func recordError(span opentracing.Span, err error) {
	span.LogFields(
		log.Bool("success", false),
		log.Error(err),
	)
}
