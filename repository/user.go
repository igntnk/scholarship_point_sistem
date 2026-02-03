package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/authorization"
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
	queries   *db.Queries
	querier   db.Querier
	txCreator db.TxCreator
}

func NewUserRepository(pool pgxv5.Tr) UserRepository {
	return &userRepository{
		queries:   db.New(pool),
		querier:   db.NewQuerier(pool),
		txCreator: db.NewTxCreator(pool),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user models.UserWithCredentials) (string, error) {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

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
	pgUUID, err := qtx.CreateUser(ctx, args)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	err = qtx.AddUserToUserGroup(ctx, pgUUID)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
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
			return models.UserWithCredentials{}, authorization.WrongPasswordErr
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
select uuid,
       name,
       second_name,
       patronymic,
       gradebook_number,
       birth_date,
       phone_number,
       email,
       user_status,
       point_amount,
       achievement_amount,
       count(*) over () as total_amount,
       achievement_proofed
from (select u.uuid,
             u.name,
             second_name,
             patronymic,
             gradebook_number,
             birth_date,
             phone_number,
             email,
             u_s.display_value                                           as user_status,
             count(distinct a.uuid)                                      as achievement_amount,
             coalesce(sum(cv.point) + sum(p_c.point_amount), 0)::numeric as point_amount,
             not bool_or(a_s.internal_value = 'unapproved')              as achievement_proofed
      from sys_user u
               join status u_s on u_s.uuid = u.status_uuid
               left join achievement a on a.user_uuid = u.uuid
               left join status a_s on a_s.uuid = a.status_uuid and a_s.type = 'achievement_status'
               left join achievement_category ac on ac.achievement_uuid = a.uuid
               left join category p_c
                         on p_c.uuid = ac.category_uuid and p_c.parent_category is null and
                            p_c.status_uuid in (select uuid
                                                from status
                                                where type = 'category_status'
                                                  and internal_value = 'active')
               left join category c_c
                         on c_c.uuid = ac.category_uuid and c_c.parent_category is not null and
                            c_c.status_uuid in (select uuid
                                                from status
                                                where type = 'category_status'
                                                  and internal_value = 'active')
               left join achievement_category_value acv on acv.achievement_uuid = a.uuid
               left join category_value cv on cv.uuid = acv.category_value_uuid and cv.category_uuid = c_c.uuid and
                                              cv.status_uuid != (select s.uuid
                                                                 from status s
                                                                 where type = 'category_value_status'
                                                                   and internal_value = 'unactive')
      where ((a_s.internal_value != 'removed' and a_s.internal_value != 'declined') or a_s.internal_value is null) %s
      group by u.uuid, u.name, second_name, patronymic, gradebook_number, birth_date, phone_number, email,
               u_s.display_value,
               c_c.point_amount
      %s) as sub
where 1 = 1
`

	searchTemplate := ""
	for i, searchWord := range searchWords {
		if searchWord == "" {
			continue
		}

		if i == 0 {
			searchTemplate = fmt.Sprintf("(%%s ilike '%%%%%s%%%%'", searchWord)
		} else {
			searchTemplate = fmt.Sprintf("%s or %%s ilike '%%%%%s%%%%'", searchTemplate, searchWord)
		}

		if i == len(searchWords)-1 {
			searchTemplate = fmt.Sprintf("%s)", searchTemplate)
		}
	}

	searchStmt := ""
	if searchTemplate != "" {
		searchStmt = fmt.Sprintf("and (%s or %s or %s or %s or %s or %s or %s)",
			searchTemplate,
			searchTemplate,
			searchTemplate,
			searchTemplate,
			searchTemplate,
			searchTemplate,
			searchTemplate,
		)
		searchStmt = fmt.Sprintf(searchStmt, "u.name",
			"second_name",
			"patronymic",
			"gradebook_number",
			"u_s.display_value",
			"phone_number",
			"email",
		)
	}

	if Valid {
		searchStmt = fmt.Sprintf("%s and u_s.display_value = 'Подтвержденный'", searchStmt)
	}

	limitStmt := ""
	if Winners {
		if !Valid {
			searchStmt = fmt.Sprintf("%s and u_s.display_value = 'Подтвержденный'", searchStmt)
		}
		limitStmt = "limit (select value::bigint from constants where name = 'available_student_grades')"
	}

	request = fmt.Sprintf(request, searchStmt, limitStmt)

	request = fmt.Sprintf("%s order by point_amount", request)

	if Limit != 0 {
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
			name                   sql.NullString
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
			&name,
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
			Name:                   name.String,
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
	err = r.queries.MakeUserVerified(context, pgUUID)
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
	err = r.queries.MakeUserUnverified(context, pgUUID)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	return nil
}
