package service

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category requests.CreateCategory) (uuid string, err error)
	GetCategoryByUuid(ctx context.Context, uuid string) (models.Category, error)
	GetParentCategoriesWithPagination(ctx context.Context, limit, offset int) ([]models.Category, int, error)
	GetParentCategories(ctx context.Context) ([]models.Category, error)
	GetChildCategories(ctx context.Context, uuid string) ([]responses.Category, error)
	DeleteCategory(ctx context.Context, uuid string) error
	UpdateCategory(ctx context.Context, category requests.UpdateCategory) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s categoryService) CreateCategory(
	ctx context.Context,
	category requests.CreateCategory,
) (
	uuid string,
	err error,
) {

	if category.ParentUuid != "" {
		return s.categoryRepo.CreateSubCategory(ctx, category.Name, category.ParentUuid, category.Values)
	}

	return s.categoryRepo.CreateCategory(ctx, category.Name, category.Points)
}

func (s categoryService) GetCategoryByUuid(
	ctx context.Context,
	uuid string,
) (
	category models.Category,
	err error,
) {
	pgUUid := pgtype.UUID{}
	if err = pgUUid.Scan(uuid); err != nil {
		return models.Category{}, errors.Join(err, parsing.InputDataErr)
	}

	dbCategory, err := s.categoryRepo.GetCategoryByUUID(ctx, pgUUid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Category{}, validation.NoDataFoundErr
		}
		return models.Category{}, errors.Join(err, unexpected.RequestErr)
	}

	dbFloat, err := dbCategory.PointAmount.Float64Value()
	if err != nil {
		return models.Category{}, errors.Join(err, parsing.OutputDataErr)
	}
	return models.Category{
		UUID:   dbCategory.Uuid.String(),
		Name:   dbCategory.Name,
		Points: float32(dbFloat.Float64),
	}, nil
}

func (s categoryService) GetParentCategoriesWithPagination(ctx context.Context, limit, offset int) ([]models.Category, int, error) {
	dbCategories, err := s.categoryRepo.GetParentCategoriesWithPagination(ctx, db.ListParentCategoriesWithPaginationParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Category{}, 0, nil
		}
		return []models.Category{}, 0, errors.Join(err, unexpected.RequestErr)
	}

	categories := make([]models.Category, len(dbCategories))
	totalAmount := 0
	for i, dbCategory := range dbCategories {
		totalAmount = int(dbCategory.TotalAmount)
		dbFloat, err := dbCategory.PointAmount.Float64Value()
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		categories[i] = models.Category{
			UUID:   dbCategory.Uuid.String(),
			Name:   dbCategory.Name,
			Points: float32(dbFloat.Float64),
		}
	}

	return categories, totalAmount, nil
}

func (s categoryService) GetParentCategories(ctx context.Context) ([]models.Category, error) {
	dbCategories, err := s.categoryRepo.GetParentCategories(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Category{}, nil
		}
		return nil, errors.Join(err, unexpected.RequestErr)
	}

	categories := make([]models.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		dbFloat, err := dbCategory.PointAmount.Float64Value()
		if err != nil {
			return nil, errors.Join(err, parsing.OutputDataErr)
		}
		categories[i] = models.Category{
			UUID:   dbCategory.Uuid.String(),
			Name:   dbCategory.Name,
			Points: float32(dbFloat.Float64),
		}
	}

	return categories, nil
}

func (s categoryService) GetChildCategories(ctx context.Context, uuid string) ([]responses.Category, error) {
	modelCategories, err := s.categoryRepo.GetChildCategories(ctx, uuid)
	if err != nil {
		return nil, err
	}

	resp := make([]responses.Category, len(modelCategories))
	for i, modelCategory := range modelCategories {

		catVals := make([]responses.CategoryValues, len(modelCategory.Values))
		for i, catVal := range modelCategory.Values {
			catVals[i] = responses.CategoryValues{
				Name:   catVal.Name,
				Points: catVal.Points,
			}
		}

		resp[i] = responses.Category{
			UUID:   modelCategory.UUID,
			Name:   modelCategory.Name,
			Values: catVals,
		}
	}

	return resp, nil
}

func (s categoryService) DeleteCategory(ctx context.Context, uuid string) error {
	pgUUid := pgtype.UUID{}
	if err := pgUUid.Scan(uuid); err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}
	err := s.categoryRepo.CheckCategoryExistsByUUID(ctx, pgUUid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}
		return err
	}

	return s.categoryRepo.DeleteCategory(ctx, pgUUid)
}

func (s categoryService) UpdateCategory(ctx context.Context, category requests.UpdateCategory) error {
	pgUUid := pgtype.UUID{}
	if err := pgUUid.Scan(category.UUID); err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	err := s.categoryRepo.CheckCategoryExistsByUUID(ctx, pgUUid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}
		return err
	}

	if len(category.Values) != 0 {
		return s.categoryRepo.UpdateSubCategory(ctx, category.UUID, category.Name, category.Values)
	}

	return s.categoryRepo.UpdateCategory(ctx, category.UUID, category.Name, category.Points)
}
