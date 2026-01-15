package repository

import (
	"context"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryRepository interface {
	CreateCategory(context.Context, db.CreateCategoryParams) (pgtype.UUID, error)
	CheckCategoryExistsByNameAndParentNull(context.Context, string) error
	CheckCategoryExistsByNameAndParentUUID(ctx context.Context, args db.GetCategoryByNameAndParentUUIDParams) error
	CheckCategoryExistsByUUID(context.Context, pgtype.UUID) error
	GetCategoryByUUID(context.Context, pgtype.UUID) (db.GetCategoryByUUIDRow, error)
	GetCategoriesWithPagination(context.Context, db.ListCategoriesWithPaginationParams) ([]db.ListCategoriesWithPaginationRow, error)
	GetCategories(ctx context.Context) ([]db.ListCategoriesRow, error)
	DeleteCategory(context.Context, pgtype.UUID) error
	UpdateCategory(context.Context, db.UpdateCategoryParams) error
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

func (r *categoryRepository) GetCategoriesWithPagination(ctx context.Context, args db.ListCategoriesWithPaginationParams) ([]db.ListCategoriesWithPaginationRow, error) {
	return r.queries.ListCategoriesWithPagination(ctx, args)
}

func (r *categoryRepository) GetCategories(ctx context.Context) ([]db.ListCategoriesRow, error) {
	return r.queries.ListCategories(ctx)
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, uuid pgtype.UUID) error {
	return r.queries.DeleteCategory(ctx, uuid)
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, args db.UpdateCategoryParams) error {
	return r.queries.UpdateCategory(ctx, args)
}
