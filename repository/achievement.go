package repository

import (
	"context"
	"errors"
	"github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
)

type AchievementRepository interface {
	GetSimpleUserAchievementByUUID(ctx context.Context, uuid string) (models.SimpleAchievement, error)
	GetAchievementCategories(ctx context.Context, uuid string) ([]models.Category, error)
	GetUserAchievements(ctx context.Context, uuid string) ([]models.SimpleAchievement, error)
	GetUserAchievementsWithPagination(ctx context.Context, uuid string, limit, offset int) ([]models.SimpleAchievement, int, error)
	MakeAchievementUnapproved(ctx context.Context, uuid string) error
	MakeAchievementApproved(ctx context.Context, uuid string) error
	MakeAchievementUsed(ctx context.Context, uuid string) error
	MakeAchievementDeclined(ctx context.Context, uuid string) error
	MakeAchievementRemoved(ctx context.Context, uuid string) error
	CreateAchievementWithCategories(ctx context.Context, achievement models.Achievement) (string, error)
	UpdateAchievementDescFields(ctx context.Context, achievement models.SimpleAchievement) error
	UpdateAchievementFull(ctx context.Context, achievement models.Achievement) error
}

type achievementRepository struct {
	queries   *db.Queries
	txCreator db.TxCreator
}

func NewAchievementRepository(pool pgxv5.Tr) AchievementRepository {
	return &achievementRepository{
		queries:   db.New(pool),
		txCreator: db.NewTxCreator(pool),
	}
}

func (r *achievementRepository) GetSimpleUserAchievementByUUID(ctx context.Context, uuid string) (models.SimpleAchievement, error) {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return models.SimpleAchievement{}, errors.Join(err, validation.WrongInputErr)
	}

	dbAchievement, err := r.queries.GetSimpleUserAchievementByUUID(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.SimpleAchievement{}, validation.NoDataFoundErr
		}
		return models.SimpleAchievement{}, errors.Join(err, unexpected.RequestErr)
	}

	return models.SimpleAchievement{
		UUID:           dbAchievement.Uuid.String(),
		AttachmentLink: dbAchievement.AttachmentLink,
		Status:         dbAchievement.Status.String,
		CategoryName:   dbAchievement.CategoryName,
		Comment:        dbAchievement.Comment.String,
	}, nil
}

func (r *achievementRepository) GetAchievementCategories(ctx context.Context, uuid string) ([]models.Category, error) {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return []models.Category{}, errors.Join(err, validation.WrongInputErr)
	}

	cats, err := r.queries.GetCategoryByAchievement(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Category{}, nil
		}
		return []models.Category{}, errors.Join(err, unexpected.RequestErr)
	}

	resp := make([]models.Category, len(cats))
	for _, cat := range cats {

		pointAmountFloat, err := cat.PointAmount.Float64Value()
		if err != nil {
			return []models.Category{}, errors.Join(err, unexpected.InternalErr)
		}

		resp = append(resp, models.Category{
			UUID:               cat.Uuid.String(),
			Name:               cat.Name,
			ParentCategoryUUID: cat.ParentCategory.String(),
			PointAmount:        float32(pointAmountFloat.Float64),
			Comment:            cat.Comment.String,
			Status:             cat.StatusValue.String,
		})
	}

	return resp, nil
}

func (r *achievementRepository) GetUserAchievements(ctx context.Context, uuid string) ([]models.SimpleAchievement, error) {
	pgUserUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return []models.SimpleAchievement{}, errors.Join(err, validation.WrongInputErr)
	}

	dbAchievements, err := r.queries.GetUserAchievements(ctx, pgUserUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.SimpleAchievement{}, nil
		}
		return []models.SimpleAchievement{}, errors.Join(err, unexpected.RequestErr)
	}

	modelAchievement := make([]models.SimpleAchievement, len(dbAchievements))
	for i, dbAchievement := range dbAchievements {
		modelAchievement[i] = models.SimpleAchievement{
			UUID:           dbAchievement.Uuid.String(),
			AttachmentLink: dbAchievement.AttachmentLink,
			Status:         dbAchievement.Status.String,
			CategoryName:   dbAchievement.CategoryName,
			Comment:        dbAchievement.Comment.String,
		}
	}
	return modelAchievement, nil
}

func (r *achievementRepository) GetUserAchievementsWithPagination(ctx context.Context, uuid string, limit, offset int) ([]models.SimpleAchievement, int, error) {
	pgUserUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return []models.SimpleAchievement{}, 0, errors.Join(err, validation.WrongInputErr)
	}

	args := db.GetUserAchievementsWithPaginationParams{
		Limit:    int32(limit),
		Offset:   int32(offset),
		UserUuid: pgUserUUID,
	}
	dbAchievements, err := r.queries.GetUserAchievementsWithPagination(ctx, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.SimpleAchievement{}, 0, nil
		}
		return []models.SimpleAchievement{}, 0, errors.Join(err, unexpected.RequestErr)
	}

	modelAchievements := make([]models.SimpleAchievement, len(dbAchievements))
	totalRecords := 0
	for i, dbAchievement := range dbAchievements {
		totalRecords = int(dbAchievement.TotalRecords)
		modelAchievements[i] = models.SimpleAchievement{
			UUID:           dbAchievement.Uuid.String(),
			AttachmentLink: dbAchievement.AttachmentLink,
			Status:         dbAchievement.Status.String,
			CategoryName:   dbAchievement.CategoryName,
			Comment:        dbAchievement.Comment.String,
		}
	}
	return modelAchievements, totalRecords, nil
}

func (r *achievementRepository) MakeAchievementUnapproved(ctx context.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	err = r.queries.MakeAchievementUnapproved(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}

		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (r *achievementRepository) MakeAchievementApproved(ctx context.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	err = r.queries.MakeAchievementApproved(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}

		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (r *achievementRepository) MakeAchievementUsed(ctx context.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	err = r.queries.MakeAchievementUsed(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}

		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (r *achievementRepository) MakeAchievementRemoved(ctx context.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	err = r.queries.MakeAchievementRemoved(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}

		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (r *achievementRepository) MakeAchievementDeclined(ctx context.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	err = r.queries.MakeAchievementDeclined(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}

		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (r *achievementRepository) CreateAchievementWithCategories(ctx context.Context, achievement models.Achievement) (string, error) {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return "", errors.Join(err, unexpected.InternalErr)
	}

	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgComment, err := ParseToPgText(achievement.Comment)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	pgUserUUUD, err := ParseToPgUUID(achievement.UUID)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	achievementArg := db.CreateAchievementParams{
		Comment:        pgComment,
		AttachmentLink: achievement.AttachmentLink,
		UserUuid:       pgUserUUUD,
	}
	achievementUUID, err := qtx.CreateAchievement(ctx, achievementArg)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	achCatArg := make([]db.CreateBatchAchievementCategoryParams, len(achievement.Categories))
	for i := range achCatArg {
		cat := achievement.Categories[i]

		pgCatUUID, err := ParseToPgUUID(cat.UUID)
		if err != nil {
			return "", errors.Join(err, unexpected.RequestErr)
		}
		achCatArg[i] = db.CreateBatchAchievementCategoryParams{
			CategoryUuid:    pgCatUUID,
			AchievementUuid: achievementUUID,
		}
	}

	catBatch := qtx.CreateBatchAchievementCategory(ctx, achCatArg)
	catBatch.Exec(func(i int, err error) {
		if err != nil {
			err = errors.Join(err, unexpected.RequestErr)
		}
	})
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}
	if err = tx.Commit(ctx); err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	return achievementUUID.String(), nil
}

func (r *achievementRepository) UpdateAchievementDescFields(ctx context.Context, achievement models.SimpleAchievement) error {

	pgUUID, err := ParseToPgUUID(achievement.UUID)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}

	pgComment, err := ParseToPgText(achievement.Comment)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}

	args := db.UpdateAchievementParams{
		Comment:        pgComment,
		AttachmentLink: achievement.AttachmentLink,
		Uuid:           pgUUID,
	}
	err = r.queries.UpdateAchievement(ctx, args)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *achievementRepository) UpdateAchievementFull(ctx context.Context, achievement models.Achievement) error {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.InternalErr)
	}

	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgAchievementUUID, err := ParseToPgUUID(achievement.UUID)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}

	err = qtx.RemoveAllAchievementCategory(ctx, pgAchievementUUID)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	catArgs := make([]db.CreateBatchAchievementCategoryParams, len(achievement.Categories))
	for i := range achievement.Categories {
		cat := achievement.Categories[i]
		pgCatUUID, err := ParseToPgUUID(cat.UUID)
		if err != nil {
			return errors.Join(err, validation.WrongInputErr)
		}
		catArgs[i] = db.CreateBatchAchievementCategoryParams{
			CategoryUuid:    pgCatUUID,
			AchievementUuid: pgAchievementUUID,
		}
	}
	batch := qtx.CreateBatchAchievementCategory(ctx, catArgs)
	batch.Exec(func(i int, err error) {
		if err != nil {
			err = errors.Join(err, unexpected.RequestErr)
		}
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	pgComment, err := ParseToPgText(achievement.Comment)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}

	args := db.UpdateAchievementParams{
		Comment:        pgComment,
		AttachmentLink: achievement.AttachmentLink,
		Uuid:           pgAchievementUUID,
	}
	err = qtx.UpdateAchievement(ctx, args)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	err = qtx.MakeAchievementUnapproved(ctx, pgAchievementUUID)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}
