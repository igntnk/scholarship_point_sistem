package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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
	GetShortInfoRating(ctx context.Context, userUUID string) (responses.RatingShortInfo, error)
	ApproveUser(context *gin.Context, uuid string) error
	DeclineUser(context *gin.Context, uuid string) error
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
	request := `
select *
from (select distinct u.uuid,
                      u.name,
                      second_name,
                      patronymic,
                      gradebook_number,
                      birth_date,
                      phone_number,
                      email,
                      u_s.display_value              as user_status,
                      sum(cv.point) + c.point_amount as point_amount,
                      count(a.uuid)                  as achievement_amount,
                      count(*) over ()               as total_amount,
                      case
                          when max(case when a_s.internal_value = 'unapproved' then 1 else 0 end) = 1
                              then false
                          else true
                          end                        as all_achievement_verified
      from sys_user u
               left join achievement a on a.user_uuid = u.uuid and a.status_uuid in (select uuid
                                                                                     from status
                                                                                     where type = 'achievement_status'
                                                                                       and internal_value != 'declined')
               left join achievement_category ac on ac.achievement_uuid = a.uuid
               left join category c
                         on c.uuid = ac.category_uuid and parent_category is null and c.status_uuid in (select uuid
                                                                                                        from status
                                                                                                        where type = 'category_status'
                                                                                                          and internal_value = 'active')
               left join category c_p on c_p.uuid = ac.category_uuid and c_p.parent_category is not null and
                                         c_p.status_uuid in (select uuid
                                                             from status
                                                             where type = 'category_status'
                                                               and internal_value = 'active')
               left join achievement_category_value acv on acv.achievement_uuid = a.uuid
               left join category_value cv on cv.uuid = acv.category_value_uuid
               left join status u_s on u_s.uuid = u.status_uuid and u_s.type = 'user_status'
               left join status a_s on a_s.uuid = a.status_uuid and a_s.type = 'achievement_status'
      where c.point_amount is not null
      group by u.uuid, u.name, second_name, patronymic, gradebook_number, birth_date, phone_number, email,
               c.point_amount, u_s.display_value) as sub
where 1 = 1
`

	name := ""
	secondName := ""
	patronymic := ""
	gradebookNumber := ""
	phoneNumber := ""
	email := ""
	userStatus := ""
	for i, searchWord := range searchWords {
		if searchWord == "" {
			continue
		}

		if i == 0 {
			name = fmt.Sprintf("(name ilike '%%%s%%'", searchWord)
			secondName = fmt.Sprintf("(second_name ilike '%%%s%%'", searchWord)
			patronymic = fmt.Sprintf("(patronymic ilike '%%%s%%'", searchWord)
			gradebookNumber = fmt.Sprintf("(gradebook_number ilike '%%%s%%'", searchWord)
			phoneNumber = fmt.Sprintf("(phone_number ilike '%%%s%%'", searchWord)
			email = fmt.Sprintf("(email ilike '%%%s%%' ", searchWord)
			userStatus = fmt.Sprintf("(user_status ilike '%%%s%%'", searchWord)
		} else {
			name = fmt.Sprintf("%s or name ilike '%%%s%%'", name, searchWord)
			secondName = fmt.Sprintf("%s or second_name ilike '%%%s%%'", secondName, searchWord)
			patronymic = fmt.Sprintf("%s or patronymic ilike '%%%s%%'", patronymic, searchWord)
			gradebookNumber = fmt.Sprintf("%s or gradebook_number ilike '%%%s%%'", gradebookNumber, searchWord)
			phoneNumber = fmt.Sprintf("%s or phone_number ilike '%%%s%%'", phoneNumber, searchWord)
			email = fmt.Sprintf("%s or email ilike '%%%s%%'", email, searchWord)
			userStatus = fmt.Sprintf("%s or user_status ilike '%%%s%%'", userStatus, searchWord)
		}

		if i == len(searchWords)-1 {
			name = fmt.Sprintf("%s)", name)
			secondName = fmt.Sprintf("%s)", secondName)
			patronymic = fmt.Sprintf("%s)", patronymic)
			gradebookNumber = fmt.Sprintf("%s)", gradebookNumber)
			phoneNumber = fmt.Sprintf("%s)", phoneNumber)
			email = fmt.Sprintf("%s)", email)
			userStatus = fmt.Sprintf("%s)", userStatus)
		}
	}

	if name != "" {
		request = fmt.Sprintf("%s and (%s or %s or %s or %s or %s or %s or %s)",
			request,
			name,
			secondName,
			patronymic,
			gradebookNumber,
			phoneNumber,
			email,
			userStatus,
		)
	}

	if Valid {
		request = fmt.Sprintf("%s and user_status = 'Подтвержденный'", request)
	}

	request = fmt.Sprintf("%s order by point_amount", request)

	if Winners {
		request = fmt.Sprintf("%s limit (select value::bigint from constants where name = 'available_student_grades')", request)
	} else if Limit != 0 {
		request = fmt.Sprintf("%s limit %d offset %d", request, Limit, Offset)
	}

	rows, err := r.querier.Exec(ctx, request)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, validation.NoDataFoundErr
		}
		return nil, 0, errors.Join(err, unexpected.RequestErr)
	}

	users := make([]responses.User, 0)
	TotalAmount := 0
	for rows.Next() {
		var (
			firstName              sql.NullString
			secondName             sql.NullString
			uuid                   string
			patronymic             sql.NullString
			phoneNumber            sql.NullString
			gradebookNumber        string
			status                 string
			birthDate              sql.NullString
			email                  sql.NullString
			allAchievementVerified sql.NullBool
			pointAmount            float64
			achievementAmount      int
		)
		if err = rows.Scan(
			&uuid,
			&firstName,
			&secondName,
			&patronymic,
			&gradebookNumber,
			&birthDate,
			&phoneNumber,
			&email,
			&status,
			&pointAmount,
			&achievementAmount,
			&TotalAmount,
			&allAchievementVerified,
		); err != nil {
			return nil, 0, errors.Join(err, unexpected.RequestErr)
		}

		user := responses.User{
			UUID:                   uuid,
			Name:                   firstName.String,
			SecondName:             secondName.String,
			Patronymic:             patronymic.String,
			BirthDate:              birthDate.String,
			PhoneNumber:            phoneNumber.String,
			GradebookNumber:        gradebookNumber,
			Email:                  email.String,
			PointsAmount:           pointAmount,
			AchievementAmount:      achievementAmount,
			Valid:                  status == "Подтвержденный",
			AllAchievementVerified: allAchievementVerified.Bool,
		}

		users = append(users, user)
	}

	return users, TotalAmount, nil
}

func (r *userRepository) GetShortInfoRating(ctx context.Context, userUUID string) (responses.RatingShortInfo, error) {
	pgUUID, err := ParseToPgUUID(userUUID)
	if err != nil {
		return responses.RatingShortInfo{}, errors.Join(err, parsing.InputDataErr)
	}

	dbShortInfo, err := r.queries.GetShortRatingInfo(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return responses.RatingShortInfo{}, validation.NoDataFoundErr
		}
		return responses.RatingShortInfo{}, errors.Join(err, unexpected.RequestErr)
	}

	result := responses.RatingShortInfo{}
	for i, info := range dbShortInfo {
		if i == 0 {
			result.LeaderPoints = float64(info.PointAmount)
			if len(dbShortInfo)-1 == 0 {
				result.CurrentPoints = float64(info.PointAmount)
				result.CurrentPosition = int(info.Position)
			}
			continue
		}

		result.CurrentPoints = float64(info.PointAmount)
		result.CurrentPosition = int(info.Position)
	}

	return result, nil
}

func (r *userRepository) ApproveUser(context *gin.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}
	err = r.queries.MakeAchievementApproved(context, pgUUID)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}
func (r *userRepository) DeclineUser(context *gin.Context, uuid string) error {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}
	err = r.queries.MakeAchievementDeclined(context, pgUUID)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}
