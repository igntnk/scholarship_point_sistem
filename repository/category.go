package repository

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryRepository interface {
	CreateCategory(context.Context, db.CreateCategoryParams) (pgtype.UUID, error)
	CheckCategoryExistsByName(context.Context, string) (bool, error)
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

func (r *categoryRepository) CheckCategoryExistsByName(
	ctx context.Context,
	name string,
) (
	bool,
	error,
) {
	category, err := r.queries.GetCategoryByName(ctx, name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	return category.Uuid.Valid, nil
}
