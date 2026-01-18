package repository

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, args models.UserWithCredentials) (string, error)
	GetApprovedUserByGradeBookNumber(ctx context.Context, gradeNumber string) (db.GetApprovedUserByGradeBookNumberRow, error)
	GetSimpleUserByUUID(ctx context.Context, uuid string) (models.SimpleUser, error)
	GetSimpleUserList(ctx context.Context) ([]db.SysUser, error)
	GetSimpleUserListWithPagination(ctx context.Context, args db.GetSimpleUserListWithPaginationParams) ([]db.GetSimpleUserListWithPaginationRow, error)
	UpdateUserWithoutGradeBook(ctx context.Context, args db.UpdateUserInfoWithoutGradeBookParams) error
	UpdateUserWithGradeBook(ctx context.Context, args db.UpdateUserInfoWithGradeBookParams) error
	GetUserWithCredentialsByEmail(ctx context.Context, email string) (models.UserWithCredentials, error)
}

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(pool db.DBTX) UserRepository {
	return &userRepository{
		queries: db.New(pool),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user models.UserWithCredentials) (string, error) {
	pgBDate, err := ParseToPgDate(user.BirthDate)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	pgPatronymic, err := ParseToPgText(user.Patronymic)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	pgEmail, err := ParseToPgText(user.Email)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	pgPhoneNumber, err := ParseToPgText(user.PhoneNumber)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	pgSalt, err := ParseToPgText(user.Salt)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	pgPassword, err := ParseToPgText(user.HashedPassword)
	if err != nil {
		return "", errors.Join(err, validation.WrongInputErr)
	}

	args := db.CreateUserParams{
		Name:            user.Name,
		SecondName:      user.SecondName,
		Patronymic:      pgPatronymic,
		GradebookNumber: user.GradeBookNumber,
		BirthDate:       pgBDate,
		Email:           pgEmail,
		PhoneNumber:     pgPhoneNumber,
		Password:        pgPassword,
		Salt:            pgSalt,
	}
	pgUUID, err := r.queries.CreateUser(ctx, args)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	return pgUUID.String(), nil
}

func (r *userRepository) GetApprovedUserByGradeBookNumber(ctx context.Context, gradeNumber string) (db.GetApprovedUserByGradeBookNumberRow, error) {
	return r.queries.GetApprovedUserByGradeBookNumber(ctx, gradeNumber)
}

func (r *userRepository) GetSimpleUserByUUID(ctx context.Context, uuid string) (models.SimpleUser, error) {
	pgUuid, err := ParseToPgUUID(uuid)
	if err != nil {
		return models.SimpleUser{}, errors.Join(err, parsing.InputDataErr)
	}
	dbUser, err := r.queries.GetSimpleUserByUUID(ctx, pgUuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.SimpleUser{}, errors.Join(err, validation.NoDataFoundErr)
		}
		return models.SimpleUser{}, errors.Join(err, unexpected.RequestErr)
	}

	return models.SimpleUser{
		UUID:            dbUser.Uuid.String(),
		Name:            dbUser.Name,
		SecondName:      dbUser.SecondName,
		Patronymic:      dbUser.Patronymic.String,
		GradeBookNumber: dbUser.GradebookNumber,
		BirthDate:       dbUser.BirthDate.Time.Format(time.RFC3339),
		Email:           dbUser.Email.String,
		PhoneNumber:     dbUser.PhoneNumber.String,
	}, nil
}

func (r *userRepository) GetSimpleUserList(ctx context.Context) ([]db.SysUser, error) {
	return r.queries.GetSimpleUserList(ctx)
}

func (r *userRepository) GetSimpleUserListWithPagination(ctx context.Context, args db.GetSimpleUserListWithPaginationParams) ([]db.GetSimpleUserListWithPaginationRow, error) {
	return r.queries.GetSimpleUserListWithPagination(ctx, args)
}

func (r *userRepository) UpdateUserWithoutGradeBook(ctx context.Context, args db.UpdateUserInfoWithoutGradeBookParams) error {
	return r.queries.UpdateUserInfoWithoutGradeBook(ctx, args)
}

func (r *userRepository) UpdateUserWithGradeBook(ctx context.Context, args db.UpdateUserInfoWithGradeBookParams) error {
	return r.queries.UpdateUserInfoWithGradeBook(ctx, args)
}

func (r *userRepository) GetUserWithCredentialsByEmail(ctx context.Context, email string) (models.UserWithCredentials, error) {
	pgEmail, err := ParseToPgText(email)
	if err != nil {
		return models.UserWithCredentials{}, errors.Join(err, validation.WrongInputErr)
	}

	dbUser, err := r.queries.GetUserByEmail(ctx, pgEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserWithCredentials{}, errors.Join(err, validation.NoDataFoundErr)
		}
		return models.UserWithCredentials{}, errors.Join(err, unexpected.RequestErr)
	}

	return models.UserWithCredentials{
		UUID:            dbUser.Uuid.String(),
		Name:            dbUser.Name,
		SecondName:      dbUser.SecondName,
		Patronymic:      dbUser.Patronymic.String,
		GradeBookNumber: dbUser.GradebookNumber,
		BirthDate:       dbUser.BirthDate.Time.Format(time.RFC3339),
		Email:           dbUser.Email.String,
		PhoneNumber:     dbUser.PhoneNumber.String,
		HashedPassword:  dbUser.Password.String,
		Salt:            dbUser.Salt.String,
	}, nil
}
