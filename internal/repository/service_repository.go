package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/CzarSimon/httputil/dbutil"
	"github.com/opentracing/opentracing-go"
	tracelog "github.com/opentracing/opentracing-go/log"
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

const insertServiceQuery = `
	INSERT INTO service(
		id, 
		application, 
		location, 
		port, 
		status,
		created_at,
		updated_at
	) VALUES (
		?,
		?,
		?,
		?,
		?,
		?,
		?
	)`

func (r *serviceRepo) Save(ctx context.Context, svc dto.Service) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serviceRepo.Save")
	defer span.Finish()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("failed to create transaction. %w", err)
		recordError(span, err)
		dbutil.Rollback(tx)
		return err
	}

	s := dto.Service{}
	err = tx.QueryRowContext(ctx, findQuery, svc.ID).Scan(&s.ID, &s.Application, &s.Location, &s.Port, &s.Status)
	if err == nil {
		fmt.Println("TODO: update service", s.ID)
		return nil // r.updateService(ctx, tx, svc)
	} else if err != nil && err != sql.ErrNoRows {
		err = fmt.Errorf("failed to query for existing service. %w", err)
		recordError(span, err)
		dbutil.Rollback(tx)
		return err
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, insertServiceQuery, svc.ID, svc.Application, svc.Location, svc.Port, svc.Status, now, now)
	if err != nil {
		err = fmt.Errorf("failed to insert new service. %w", err)
		recordError(span, err)
		dbutil.Rollback(tx)
		return err
	}

	span.LogFields(tracelog.Bool("success", true))
	return tx.Commit()
}

const findQuery = `
	SELECT 
		id, 
		application, 
		location, 
		port, 
		status 
	FROM service
	WHERE 
		id = ?`

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
	defer rows.Close()

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

	span.LogFields(tracelog.Bool("success", true))
	return services, nil
}

func recordError(span opentracing.Span, err error) {
	span.LogFields(
		tracelog.Bool("success", false),
		tracelog.Error(err),
	)
}
