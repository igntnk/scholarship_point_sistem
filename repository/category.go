package repository

import (
	"context"
	"errors"
	"github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, name string, pointAmount float32) (string, error)
	CreateSubCategory(ctx context.Context, name, parentUUID string, catVal []requests.CategoryValues) (string, error)
	CheckCategoryExistsByNameAndParentNull(context.Context, string) error
	CheckCategoryExistsByNameAndParentUUID(ctx context.Context, args db.GetCategoryByNameAndParentUUIDParams) error
	CheckCategoryExistsByUUID(context.Context, pgtype.UUID) error
	GetCategoryByUUID(context.Context, pgtype.UUID) (db.GetCategoryByUUIDRow, error)
	GetParentCategoriesWithPagination(context.Context, db.ListParentCategoriesWithPaginationParams) ([]db.ListParentCategoriesWithPaginationRow, error)
	GetParentCategories(ctx context.Context) ([]db.ListParentCategoriesRow, error)
	DeleteCategory(context.Context, pgtype.UUID) error
	UpdateCategory(ctx context.Context, uuid, name string, pointAmount float32) error
	UpdateSubCategory(ctx context.Context, uuid, name string, catVal []requests.CategoryValues) error
	GetChildCategories(ctx context.Context, uuid string) ([]models.Category, error)
}

type categoryRepository struct {
	queries   *db.Queries
	txCreator db.TxCreator
}

func NewCategoryRepository(pool pgxv5.Tr) CategoryRepository {
	return &categoryRepository{
		queries:   db.New(pool),
		txCreator: db.NewTxCreator(pool),
	}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, name string, pointAmount float32) (string, error) {
	pgPointAmount, err := ParseToPgNumeric(pointAmount)
	if err != nil {
		return "", errors.Join(err, parsing.InputDataErr)
	}

	catUUID, err := r.queries.CreateCategory(ctx, db.CreateCategoryParams{
		Name:        name,
		PointAmount: pgPointAmount,
	})
	if err != nil {
		return "", errors.Join(err, parsing.InputDataErr)
	}

	return catUUID.String(), nil
}

func (r *categoryRepository) CreateSubCategory(ctx context.Context, name, parentUUID string, catVal []requests.CategoryValues) (string, error) {
	pgParentUUID, err := ParseToPgUUID(parentUUID)
	if err != nil {
		return "", errors.Join(err, parsing.InputDataErr)
	}

	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	catUUID, err := qtx.CreateCategory(ctx, db.CreateCategoryParams{
		Name:           name,
		ParentCategory: pgParentUUID,
	})
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	args := make([]db.CreateCategoryValuesParams, len(catVal))
	for i := range catVal {
		cat := catVal[i]

		pgPoint, err := ParseToPgNumeric(cat.Points)
		if err != nil {
			return "", errors.Join(err, parsing.InputDataErr)
		}

		args[i] = db.CreateCategoryValuesParams{
			Name:         cat.Name,
			CategoryUuid: catUUID,
			Point:        pgPoint,
		}
	}

	catBatch := qtx.CreateCategoryValues(ctx, args)
	catBatch.Exec(func(i int, errL error) {
		if err != nil {
			err = errors.Join(err, errL)
		}
	})
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	return catUUID.String(), nil
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

func (r *categoryRepository) UpdateCategory(ctx context.Context, uuid, name string, pointAmount float32) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	// Todo сделать нормальные статусы
	pgDisplayValue, err := ParseToPgText("Активно")
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	pgPointAmount, err := ParseToPgNumeric(pointAmount)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	err = r.queries.UpdateCategory(ctx, db.UpdateCategoryParams{
		Name:         name,
		PointAmount:  pgPointAmount,
		DisplayValue: pgDisplayValue,
		Uuid:         pgUUID,
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (r *categoryRepository) UpdateSubCategory(ctx context.Context, uuid, name string, catVal []requests.CategoryValues) error {
	catUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	err = qtx.UpdateCategory(ctx, db.UpdateCategoryParams{
		Name: name,
		Uuid: catUUID,
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = qtx.DeleteCategoryValues(ctx, catUUID); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	args := make([]db.CreateCategoryValuesParams, len(catVal))
	for i := range catVal {
		cat := catVal[i]

		pgPoint, err := ParseToPgNumeric(cat.Points)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}

		args[i] = db.CreateCategoryValuesParams{
			Name:         cat.Name,
			CategoryUuid: catUUID,
			Point:        pgPoint,
		}
	}

	catBatch := qtx.CreateCategoryValues(ctx, args)
	catBatch.Exec(func(i int, err error) {
		if err != nil {
			err = errors.Join(err, unexpected.RequestErr)
		}
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
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

	result := make([]models.Category, 0, len(dbCategories))
	mapResult := make(map[string]models.Category)
	for _, dbCategory := range dbCategories {
		pointAmount, err := dbCategory.Point.Float64Value()
		if err != nil {
			return nil, errors.Join(err, unexpected.InternalErr)
		}

		cat := models.Category{}
		catVal := models.CategoryValues{
			Name:   dbCategory.ValueName,
			Points: float32(pointAmount.Float64),
		}
		var ok bool
		if cat, ok = mapResult[dbCategory.Uuid.String()]; ok {
			cat.Values = append(cat.Values, catVal)
			mapResult[dbCategory.Uuid.String()] = cat
			continue
		}
		cat = models.Category{
			UUID: pgUUID.String(),
			Name: dbCategory.CategoryName,
		}
		cat.Values = append(cat.Values, catVal)

		mapResult[dbCategory.Uuid.String()] = cat
	}

	for _, cat := range mapResult {
		result = append(result, cat)
	}

	return result, nil
}
