package repository

import (
	"database/sql"

	"github.com/rtcheap/dto"
)

// ServiceRepository storage interface for service metadata.
type ServiceRepository interface {
	Save(svc dto.Service) error
	Find(id string) (dto.Service, error)
	FindByApplication(application string) ([]dto.Service, error)
}

type serviceRepo struct {
	db *sql.DB
}
