package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type UserService interface {
	CreateUser(ctx context.Context, user requests.CreateUser) (uuid string, err error)
	GetSimpleUserList(ctx context.Context) ([]models.SimpleUser, error)
	GetSimpleUserListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleUser, int, error)
	GetSimpleUserByUUID(ctx context.Context, uuid string) (models.SimpleUser, error)
	UpdateUser(ctx context.Context, user requests.UpdateUser) error
	ApproveUser(context *gin.Context, uuid string) error
	DeclineUser(context *gin.Context, uuid string) error
}

type userService struct {
	userRepo        repository.UserRepository
	passwordManager PasswordManager
}

func NewUserService(
	userRepo repository.UserRepository,
	passwordManager PasswordManager,
) UserService {
	return &userService{
		userRepo:        userRepo,
		passwordManager: passwordManager,
	}
}

func (s *userService) CreateUser(ctx context.Context, user requests.CreateUser) (uuid string, err error) {

	_, err = s.userRepo.GetApprovedUserByGradeBookNumber(ctx, user.GradebookNumber)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid, errors.Join(err, unexpected.RequestErr)
		}
	} else {
		return uuid, validation.RecordAlreadyExistsErr
	}

	if user.Name == "" {
		return "", errors.Join(validation.WrongInputErr, errors.New("Отсуствует имя пользователя"))
	}

	if user.GradebookNumber == "" {
		return "", errors.Join(validation.WrongInputErr, errors.New("Отсуствует номер зачетной книжки"))
	}

	if user.Email == "" {
		return "", errors.Join(validation.WrongInputErr, errors.New("Отсутсвует почта"))
	}

	if err = s.passwordManager.ValidatePassword(user.Password); err != nil {
		return "", err
	}

	hashedPassword, salt, err := s.passwordManager.HashPassword(user.Password)
	if err != nil {
		return "", err
	}

	return s.userRepo.CreateUser(ctx, models.UserWithCredentials{
		Name:            user.Name,
		SecondName:      user.SecondName,
		Patronymic:      user.Patronymic,
		GradeBookNumber: user.GradebookNumber,
		BirthDate:       user.BirthDate,
		Email:           user.Email,
		PhoneNumber:     user.PhoneNumber,
		HashedPassword:  hashedPassword,
		Salt:            salt,
	})
}

func (s *userService) GetSimpleUserList(ctx context.Context) ([]models.SimpleUser, error) {
	dbUsers, err := s.userRepo.GetSimpleUserList(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, validation.NoDataFoundErr
		}
		return nil, errors.Join(err, unexpected.RequestErr)
	}

	users := make([]models.SimpleUser, len(dbUsers))
	for i, u := range dbUsers {
		users[i] = models.SimpleUser{
			UUID:            u.Uuid.String(),
			Name:            u.Name,
			SecondName:      u.SecondName,
			Patronymic:      u.Patronymic.String,
			GradeBookNumber: u.GradebookNumber,
			BirthDate:       u.BirthDate.Time.Format(time.RFC3339),
			Email:           u.Email.String,
			PhoneNumber:     u.PhoneNumber.String,
		}
	}
	return users, nil
}

func (s *userService) GetSimpleUserListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleUser, int, error) {

	args := db.GetSimpleUserListWithPaginationParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	dbUsers, err := s.userRepo.GetSimpleUserListWithPagination(ctx, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, validation.NoDataFoundErr
		}
		return nil, 0, errors.Join(err, unexpected.RequestErr)
	}

	var totalRecords int
	users := make([]models.SimpleUser, len(dbUsers))
	for i, u := range dbUsers {
		totalRecords = int(u.TotalAmount)
		users[i] = models.SimpleUser{
			UUID:            u.Uuid.String(),
			Name:            u.Name,
			SecondName:      u.SecondName,
			Patronymic:      u.Patronymic.String,
			GradeBookNumber: u.GradebookNumber,
			BirthDate:       u.BirthDate.Time.Format(time.RFC3339),
			Email:           u.Email.String,
			PhoneNumber:     u.PhoneNumber.String,
		}
	}

	return users, totalRecords, nil
}

func (s *userService) GetSimpleUserByUUID(ctx context.Context, uuid string) (models.SimpleUser, error) {
	return s.userRepo.GetSimpleUserByUUID(ctx, uuid)
}

func (s *userService) UpdateUser(ctx context.Context, user requests.UpdateUser) error {

	pgUuid := pgtype.UUID{}
	if err := pgUuid.Scan(user.UUID); err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	existingUser, err := s.userRepo.GetSimpleUserByUUID(ctx, user.UUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}
		return errors.Join(err, unexpected.RequestErr)
	}

	pgPatronymic, err := repository.ParseToPgText(user.Patronymic)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	pgBirthDate, err := repository.ParseToPgDate(user.BirthDate)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	pgPhoneNumber, err := repository.ParseToPgText(user.PhoneNumber)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	pgEmail, err := repository.ParseToPgText(user.Email)
	if err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}

	// Изменение Gradebook приводит к сбросу статуса пользователя
	if user.GradebookNumber != existingUser.GradeBookNumber {
		args := db.UpdateUserInfoWithGradeBookParams{
			Name:            user.Name,
			SecondName:      user.SecondName,
			Patronymic:      pgPatronymic,
			BirthDate:       pgBirthDate,
			PhoneNumber:     pgPhoneNumber,
			Email:           pgEmail,
			GradebookNumber: user.GradebookNumber,
			Uuid:            pgUuid,
		}
		if err = s.userRepo.UpdateUserWithGradeBook(ctx, args); err != nil {
			return errors.Join(err, unexpected.RequestErr)
		}
		return nil
	}

	args := db.UpdateUserInfoWithoutGradeBookParams{
		Name:        user.Name,
		SecondName:  user.SecondName,
		Patronymic:  pgPatronymic,
		BirthDate:   pgBirthDate,
		PhoneNumber: pgPhoneNumber,
		Email:       pgEmail,
		Uuid:        pgUuid,
	}
	if err = s.userRepo.UpdateUserWithoutGradeBook(ctx, args); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}

func (s *userService) ApproveUser(context *gin.Context, uuid string) error {
	return s.userRepo.ApproveUser(context, uuid)
}
func (s *userService) DeclineUser(context *gin.Context, uuid string) error {
	return s.userRepo.DeclineUser(context, uuid)
}
