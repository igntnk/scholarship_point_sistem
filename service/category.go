package service

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
)

type CategoryService interface {
	CreateCategory(
		ctx context.Context,
		category requests.CreateCategory,
	) (
		uuid string,
		err error,
	)
	GetCategoryByUuid(ctx context.Context, uuid string) (models.Category, error)
	GetCategoriesWithPagination(ctx context.Context, limit, offset int) ([]models.Category, int, error)
	GetCategories(ctx context.Context) ([]models.Category, error)
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

	pgParentUuid := pgtype.UUID{}
	if category.ParentUuid == "" {
		pgParentUuid.Valid = false
		err = s.categoryRepo.CheckCategoryExistsByNameAndParentNull(ctx, category.Name)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return "", unexpected.RequestErr
			}
		} else {
			return "", validation.RecordAlreadyExistsErr
		}
	} else {
		if err = pgParentUuid.Scan(category.ParentUuid); err != nil {
			return "", err
		}
		err = s.categoryRepo.CheckCategoryExistsByNameAndParentUUID(ctx, db.GetCategoryByNameAndParentUUIDParams{
			Name:           category.Name,
			ParentCategory: pgParentUuid,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return uuid, errors.Join(err, unexpected.RequestErr)
			}
		} else {
			return "", validation.RecordAlreadyExistsErr
		}
	}

	pgPointAmount := pgtype.Numeric{}
	if err = pgPointAmount.Scan(strconv.FormatFloat(float64(category.PointAmount), 'g', -1, 32)); err != nil {
		return uuid, errors.Join(err, parsing.InputDataErr)
	}

	args := db.CreateCategoryParams{
		Name:           category.Name,
		PointAmount:    pgPointAmount,
		ParentCategory: pgParentUuid,
		Comment: pgtype.Text{
			String: category.Comment,
			Valid:  true,
		},
	}

	pgUuid, err := s.categoryRepo.CreateCategory(ctx, args)
	if err != nil {
		return uuid, errors.Join(err, unexpected.RequestErr)
	}

	uuid = pgUuid.String()
	return
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
		UUID:               dbCategory.Uuid.String(),
		Name:               dbCategory.Name,
		ParentCategoryUUID: dbCategory.ParentCategory.String(),
		PointAmount:        float32(dbFloat.Float64),
		Comment:            dbCategory.Comment.String,
		Status:             dbCategory.DisplayValue.String,
	}, nil
}

func (s categoryService) GetCategoriesWithPagination(ctx context.Context, limit, offset int) ([]models.Category, int, error) {
	dbCategories, err := s.categoryRepo.GetCategoriesWithPagination(ctx, db.ListCategoriesWithPaginationParams{
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
			UUID:               dbCategory.Uuid.String(),
			Name:               dbCategory.Name,
			ParentCategoryUUID: dbCategory.ParentCategory.String(),
			PointAmount:        float32(dbFloat.Float64),
			Comment:            dbCategory.Comment.String,
			Status:             dbCategory.DisplayValue.String,
		}
	}

	return categories, totalAmount, nil
}

func (s categoryService) GetCategories(ctx context.Context) ([]models.Category, error) {
	dbCategories, err := s.categoryRepo.GetCategories(ctx)
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
			UUID:               dbCategory.Uuid.String(),
			Name:               dbCategory.Name,
			ParentCategoryUUID: dbCategory.ParentCategory.String(),
			PointAmount:        float32(dbFloat.Float64),
			Comment:            dbCategory.Comment.String,
			Status:             dbCategory.DisplayValue.String,
		}
	}

	return categories, nil
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

	pgPointAmount := pgtype.Numeric{}
	if err := pgPointAmount.Scan(strconv.FormatFloat(float64(category.PointAmount), 'g', -1, 32)); err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	pgComment := pgtype.Text{
		String: category.Comment,
		Valid:  true,
	}
	if category.Comment == "" {
		pgComment.Valid = false
	}

	args := db.UpdateCategoryParams{
		Name:        category.Name,
		PointAmount: pgPointAmount,
		Comment:     pgComment,
		DisplayValue: pgtype.Text{
			String: category.Status,
			Valid:  true,
		},
		Uuid: pgUUid,
	}
	return s.categoryRepo.UpdateCategory(ctx, args)
}
