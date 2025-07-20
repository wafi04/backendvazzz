package method

import (
	"context"

	"github.com/wafi04/backendvazzz/pkg/types"
)

type Service struct {
	Repo *MethodRepository
}

func NewMethodService(Repo *MethodRepository) *Service {
	return &Service{
		Repo: Repo,
	}
}

func (service *Service) Create(c context.Context, data types.CreateMethodData) (*types.MethodData, error) {
	return service.Repo.Create(c, &data)
}

func (service *Service) GetAll(c context.Context, skip, limit int, search, filterType string, active string) ([]types.MethodData, int, error) {
	return service.Repo.GetAll(c, skip, limit, search, filterType, active)

}

func (service *Service) Update(c context.Context, id int, data types.UpdateMethodData) (*types.MethodData, error) {
	return service.Repo.Update(c, id, &data)
}

func (service *Service) Delete(c context.Context, id int) error {
	return service.Repo.Delete(c, id)
}
