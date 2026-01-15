package service

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"time"
	"unicode"
	"unicode/utf8"
)

type UserService interface {
	CreateUser(ctx context.Context, user requests.CreateUser) (uuid string, err error)
	GetSimpleUserList(ctx context.Context) ([]models.SimpleUser, error)
	GetSimpleUserListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleUser, int, error)
	GetSimpleUserByUUID(ctx context.Context, uuid string) (models.SimpleUser, error)
	UpdateUser(ctx context.Context, user requests.UpdateUser) error
}

type userService struct {
	userRepo           repository.UserRepository
	passwordPepper     string
	passwordBcryptCost int
}

func NewUserService(
	userRepo repository.UserRepository,
	passwordPepper string,
	passwordBcryptCost int,
) UserService {
	return &userService{
		userRepo:           userRepo,
		passwordPepper:     passwordPepper,
		passwordBcryptCost: passwordBcryptCost,
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

	pgBDate, err := repository.ParseToPgDate(user.BirthDate)
	if err != nil {
		return "", errors.Join(errors.Join(validation.WrongInputErr, err), errors.New("Не получилось достать дату рождения"))
	}

	pgPatronymic, err := repository.ParseToPgText(user.Patronymic)
	if err != nil {
		return "", errors.Join(errors.Join(validation.WrongInputErr, err), errors.New("Не получилось достать отчество"))
	}

	pgEmail, err := repository.ParseToPgText(user.Email)
	if err != nil {
		return "", errors.Join(errors.Join(validation.WrongInputErr, err), errors.New("Не получилось достать email"))
	}

	pgPhoneNumber, err := repository.ParseToPgText(user.PhoneNumber)
	if err != nil {
		return "", errors.Join(errors.Join(validation.WrongInputErr, err), errors.New("Не получилось достать номер телефона"))
	}

	hashedPassword, err := s.ProcessPassword(user.Password)
	if err != nil {
		return "", err
	}

	pgPassword, err := repository.ParseToPgText(hashedPassword)
	if err != nil {
		return "", errors.Join(errors.Join(unexpected.InternalErr, err))
	}

	args := db.CreateUserParams{
		Name:            user.Name,
		SecondName:      user.SecondName,
		Patronymic:      pgPatronymic,
		GradebookNumber: user.GradebookNumber,
		BirthDate:       pgBDate,
		Email:           pgEmail,
		PhoneNumber:     pgPhoneNumber,
		Password:        pgPassword,
	}
	pgUuid, err := s.userRepo.CreateUser(ctx, args)
	if err != nil {
		return "", errors.Join(errors.Join(unexpected.RequestErr, err))
	}

	return pgUuid.String(), nil
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
	pgUuid := pgtype.UUID{}
	if err := pgUuid.Scan(uuid); err != nil {
		return models.SimpleUser{}, errors.Join(err, validation.WrongInputErr)
	}

	dbUser, err := s.userRepo.GetSimpleUserByUUID(ctx, pgUuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.SimpleUser{}, validation.NoDataFoundErr
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

func (s *userService) UpdateUser(ctx context.Context, user requests.UpdateUser) error {

	pgUuid := pgtype.UUID{}
	if err := pgUuid.Scan(user.UUID); err != nil {
		return errors.Join(err, validation.WrongInputErr)
	}
	existingUser, err := s.userRepo.GetSimpleUserByUUID(ctx, pgUuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return validation.NoDataFoundErr
		}
		return errors.Join(err, unexpected.RequestErr)
	}

	pgPatronymic, err := repository.ParseToPgText(existingUser.Patronymic)
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
	if user.GradebookNumber != existingUser.GradebookNumber {
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

func (s *userService) ProcessPassword(password string) (hashedPassword string, err error) {
	if err = validatePassword(password); err != nil {
		return "", errors.Join(validation.WrongInputErr, err)
	}

	password += s.passwordPepper
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.passwordBcryptCost)
	if err != nil {
		return "", errors.Join(errors.Join(validation.WrongInputErr, err), errors.New("Ошибка в обработке пароля"))
	}

	return string(hashedPasswordBytes), nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("Пароль обязателен для создания пользователя")
	}
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("Пароль должен быть не менее 8 символов")
	}
	if hasNoUppercase(password) {
		return errors.New("Пароль должен иметь заглавные буквы")
	}
	if hasNoDigits(password) {
		return errors.New("Пароль должен иметь цифры")
	}

	return nil
}

func hasNoUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func hasNoDigits(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
