package repository

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"strconv"
)

type ConstantRepository interface {
	GetGradeAmountsConstant(ctx context.Context) (int, error)
	UpdateGradeAmountsConstant(ctx context.Context, grade int) error
}

type constantRepository struct {
	queries *db.Queries
}

func NewConstantRepository(pool db.DBTX) ConstantRepository {
	return &constantRepository{
		queries: db.New(pool),
	}
}

func (r *constantRepository) GetGradeAmountsConstant(ctx context.Context) (int, error) {
	constStr, err := r.queries.GetGradesConstant(ctx)
	if err != nil {
		return 0, errors.Join(err, unexpected.RequestErr)
	}

	constInt, err := strconv.Atoi(constStr)
	if err != nil {
		return 0, errors.Join(err, unexpected.InternalErr)
	}

	return constInt, nil
}

func (r *constantRepository) UpdateGradeAmountsConstant(ctx context.Context, grade int) error {
	if err := r.queries.ChangeGradesConstant(ctx, strconv.Itoa(grade)); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}
