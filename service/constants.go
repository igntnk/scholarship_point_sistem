package service

import (
	"context"
	"github.com/igntnk/scholarship_point_system/repository"
)

type ConstantService interface {
	GetGradeAmountsConstant(ctx context.Context) (int, error)
	UpdateGradeAmountsConstant(ctx context.Context, grade int) error
}

type constantService struct {
	constantRepo repository.ConstantRepository
}

func NewConstantService(constantRepo repository.ConstantRepository) ConstantService {
	return &constantService{
		constantRepo: constantRepo,
	}
}

func (c *constantService) GetGradeAmountsConstant(ctx context.Context) (int, error) {
	return c.constantRepo.GetGradeAmountsConstant(ctx)
}

func (c *constantService) UpdateGradeAmountsConstant(ctx context.Context, grade int) error {
	return c.constantRepo.UpdateGradeAmountsConstant(ctx, grade)
}
