package service

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
)

type CategoryService interface {
	CreateCategory(
		ctx context.Context,
		name,
		comment,
		parentUuid string,
		pointAmount float32,
	) (
		uuid string,
		err error,
	)
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
	name,
	comment,
	parentUuid string,
	pointAmount float32,
) (
	uuid string,
	err error,
) {
	exists, err := s.categoryRepo.CheckCategoryExistsByName(ctx, name)
	if err != nil {
		return
	}
	if exists {
		return uuid, RecordAlreadyExistsErr
	}

	pgPointAmount := pgtype.Numeric{}
	if err = pgPointAmount.Scan(strconv.FormatFloat(float64(pointAmount), 'g', -1, 32)); err != nil {
		return uuid, errors.Join(err, ParsingDataErr)
	}

	pgParentUuid := pgtype.UUID{}
	if parentUuid == "" {
		pgParentUuid.Valid = false
	} else if err = pgParentUuid.Scan(parentUuid); err != nil {
		return "", err
	}

	args := db.CreateCategoryParams{
		Name:           name,
		PointAmount:    pgPointAmount,
		ParentCategory: pgParentUuid,
		Comment: pgtype.Text{
			String: comment,
			Valid:  true,
		},
	}

	pgUuid, err := s.categoryRepo.CreateCategory(ctx, args)
	if err != nil {
		return uuid, err
	}

	uuid = pgUuid.String()
	return
}
