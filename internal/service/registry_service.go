package service

import (
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
func (s *RegistryService) Register(svc dto.Service) error {
	return nil
}

// SetStatus records the status of a given service.
func (s *RegistryService) SetStatus(id string, status dto.ServiceStatus) error {
	return nil
}

// FindApplicationServices looks up all serices for an application.
func (s *RegistryService) FindApplicationServices(application string, includeUnhealthy bool) ([]dto.Service, error) {
	return nil, nil
}
