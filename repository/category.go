package repository

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryRepository interface {
	CreateCategory(context.Context, db.CreateCategoryParams) (pgtype.UUID, error)
	CheckCategoryExistsByNameAndParentNull(context.Context, string) error
	CheckCategoryExistsByNameAndParentUUID(ctx context.Context, args db.GetCategoryByNameAndParentUUIDParams) error
	CheckCategoryExistsByUUID(context.Context, pgtype.UUID) error
	GetCategoryByUUID(context.Context, pgtype.UUID) (db.GetCategoryByUUIDRow, error)
	GetParentCategoriesWithPagination(context.Context, db.ListParentCategoriesWithPaginationParams) ([]db.ListParentCategoriesWithPaginationRow, error)
	GetParentCategories(ctx context.Context) ([]db.ListParentCategoriesRow, error)
	DeleteCategory(context.Context, pgtype.UUID) error
	UpdateCategory(context.Context, db.UpdateCategoryParams) error
	GetChildCategories(ctx context.Context, uuid string) ([]models.Category, error)
}

type categoryRepository struct {
	queries *db.Queries
}

func NewCategoryRepository(pool db.DBTX) CategoryRepository {
	return &categoryRepository{
		queries: db.New(pool),
	}
}

func (r *categoryRepository) CreateCategory(
	ctx context.Context,
	args db.CreateCategoryParams,
) (
	uuid pgtype.UUID,
	err error,
) {
	return r.queries.CreateCategory(ctx, args)
}

func (r *categoryRepository) CheckCategoryExistsByNameAndParentNull(ctx context.Context, name string) error {
	_, err := r.queries.GetCategoryByNameAndParentNull(ctx, name)
	return err
}

func (r *categoryRepository) CheckCategoryExistsByNameAndParentUUID(ctx context.Context, args db.GetCategoryByNameAndParentUUIDParams) error {
	_, err := r.queries.GetCategoryByNameAndParentUUID(ctx, args)
	return err
}

func (r *categoryRepository) CheckCategoryExistsByUUID(ctx context.Context, uuid pgtype.UUID) error {
	_, err := r.queries.GetCategoryByUUID(ctx, uuid)
	return err
}

func (r *categoryRepository) GetCategoryByUUID(ctx context.Context, uuid pgtype.UUID) (db.GetCategoryByUUIDRow, error) {
	return r.queries.GetCategoryByUUID(ctx, uuid)
}

func (r *categoryRepository) GetParentCategoriesWithPagination(ctx context.Context, args db.ListParentCategoriesWithPaginationParams) ([]db.ListParentCategoriesWithPaginationRow, error) {
	return r.queries.ListParentCategoriesWithPagination(ctx, args)
}

func (r *categoryRepository) GetParentCategories(ctx context.Context) ([]db.ListParentCategoriesRow, error) {
	return r.queries.ListParentCategories(ctx)
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, uuid pgtype.UUID) error {
	return r.queries.DeleteCategory(ctx, uuid)
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, args db.UpdateCategoryParams) error {
	return r.queries.UpdateCategory(ctx, args)
}

func (r *categoryRepository) GetChildCategories(ctx context.Context, uuid string) ([]models.Category, error) {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return nil, errors.Join(err, validation.WrongInputErr)
	}
	dbCategories, err := r.queries.GetChildCategories(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Category{}, nil
		}
		return nil, errors.Join(err, unexpected.RequestErr)
	}

	result := make([]models.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		pointAmount, err := dbCategory.PointAmount.Float64Value()
		if err != nil {
			return nil, errors.Join(err, unexpected.InternalErr)
		}
		result[i] = models.Category{
			UUID:               pgUUID.String(),
			Name:               dbCategory.Name,
			ParentCategoryUUID: dbCategory.ParentCategory.String(),
			PointAmount:        float32(pointAmount.Float64),
			Comment:            dbCategory.Comment.String,
		}
	}

	return result, nil
}
