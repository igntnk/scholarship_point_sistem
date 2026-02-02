package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
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
	GetRating(ctx context.Context, searchWords []string, Valid bool, Winners bool, Limit int, Offset int) ([]responses.User, int, error)
}

type userRepository struct {
	queries *db.Queries
	querier db.Querier
}

func NewUserRepository(pool db.DBTX) UserRepository {
	return &userRepository{
		queries: db.New(pool),
		querier: db.NewQuerier(pool),
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

func (r *userRepository) GetRating(
	ctx context.Context,
	searchWords []string,
	Valid bool,
	Winners bool,
	Limit int,
	Offset int,
) (
	[]responses.User,
	int,
	error,
) {
	sql := `
select *
from (select distinct u.uuid,
                      u.name              as first_name,
                      second_name,
                      patronymic,
                      gradebook_number,
                      birth_date,
                      phone_number,
                      email,
                      s.display_value,
                      sum(c.point_amount) as user_point_amount,
                      count(a.uuid)       as achievement_amount,
                      count(*) over ()    as total_amount
      from sys_user u
               left join user_achievement ua on ua.user_uuid = u.uuid
               left join achievement a on a.uuid = ua.achievement_uuid and a.status_uuid in (select uuid
                                                                                             from status
                                                                                             where type = 'achievement_status'
                                                                                               and internal_value != 'declined')
               left join achievement_category ac on ac.achievement_uuid = a.uuid
               left join category c on c.uuid = ac.category_uuid and c.status_uuid in (select uuid
                                                                                       from status
                                                                                       where type = 'category_status'
                                                                                         and internal_value = 'active')
               left join status s on s.uuid = u.status_uuid and s.type = 'user_status'
      group by u.uuid, first_name, second_name, patronymic, gradebook_number, birth_date, phone_number, email, display_value) as sub
`
	sql = fmt.Sprintf("%s where 1=1", sql)

	searchString := ""
	for i, searchWord := range searchWords {
		if searchWord == "" {
			continue
		}
		if i == 0 {
			searchString += fmt.Sprintf("'%s'", searchWord)
			continue
		}
		searchString += fmt.Sprintf(",'%s'", searchWord)

	}

	if searchString != "" {
		sql = fmt.Sprintf("%s and first_name in (%s) or second_name in (%s) or patronymic in (%s) or gradebook_number in (%s) or phone_number in (%s) or email in (%s)", sql, searchString, searchString, searchString, searchString, searchString, searchString)
	}

	if Valid {
		sql = fmt.Sprintf("%s and status = 'Подтвержденный", sql)
	}

	if Winners {
		sql = fmt.Sprintf("%s limit (select value from constants where name = 'available_student_grades')'", sql)
	} else if Limit != 0 {
		sql = fmt.Sprintf("%s limit %d offset %d", sql, Limit, Offset)
	}

	sql = fmt.Sprintf("%s order by user_point_amount", sql)

	rows, err := r.querier.Exec(ctx, sql)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, validation.NoDataFoundErr
		}
		return nil, 0, errors.Join(err, unexpected.RequestErr)
	}

	users := make([]responses.User, 0)
	TotalAmount := 0
	for rows.Next() {
		row := map[string]any{}
		if err = rows.Scan(&row); err != nil {
			return nil, 0, errors.Join(err, unexpected.RequestErr)
		}

		firstName, err := decode[string](row["first_name"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		uuid, err := decode[string](row["uuid"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		secondName, err := decode[string](row["second_name"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		patronymic, err := decode[string](row["patronymic"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		phoneNumber, err := decode[string](row["phone_number"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		gradebookNumber, err := decode[string](row["gradebook_number"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		birthDate, err := decode[string](row["birth_date"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		email, err := decode[string](row["email"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}
		pointAmount, err := decode[float64](row["user_point_amount"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}

		achievementAmount, err := decode[int](row["achievement_amount"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}

		TotalAmount, err = decode[int](row["total_amount"])
		if err != nil {
			return nil, 0, errors.Join(err, parsing.OutputDataErr)
		}

		user := responses.User{
			UUID:                   uuid,
			Name:                   firstName,
			SecondName:             secondName,
			Patronymic:             patronymic,
			BirthDate:              birthDate,
			PhoneNumber:            phoneNumber,
			GradebookNumber:        gradebookNumber,
			Email:                  email,
			PointsAmount:           pointAmount,
			AchievementAmount:      achievementAmount,
			Valid:                  false,
			AllAchievementVerified: false,
		}

		users = append(users, user)
	}

	return users, TotalAmount, nil
}
