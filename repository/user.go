package repository

import (
	"context"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	CreateUser(ctx context.Context, args db.CreateUserParams) (pgtype.UUID, error)
	GetApprovedUserByGradeBookNumber(ctx context.Context, gradeNumber string) (db.GetApprovedUserByGradeBookNumberRow, error)
	GetSimpleUserByUUID(ctx context.Context, uuid pgtype.UUID) (db.SysUser, error)
	GetSimpleUserList(ctx context.Context) ([]db.SysUser, error)
	GetSimpleUserListWithPagination(ctx context.Context, args db.GetSimpleUserListWithPaginationParams) ([]db.GetSimpleUserListWithPaginationRow, error)
	UpdateUserWithoutGradeBook(ctx context.Context, args db.UpdateUserInfoWithoutGradeBookParams) error
	UpdateUserWithGradeBook(ctx context.Context, args db.UpdateUserInfoWithGradeBookParams) error
}

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(pool db.DBTX) UserRepository {
	return &userRepository{
		queries: db.New(pool),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, args db.CreateUserParams) (pgtype.UUID, error) {
	return r.queries.CreateUser(ctx, args)
}

func (r *userRepository) GetApprovedUserByGradeBookNumber(ctx context.Context, gradeNumber string) (db.GetApprovedUserByGradeBookNumberRow, error) {
	return r.queries.GetApprovedUserByGradeBookNumber(ctx, gradeNumber)
}

func (r *userRepository) GetSimpleUserByUUID(ctx context.Context, uuid pgtype.UUID) (db.SysUser, error) {
	return r.queries.GetSimpleUserByUUID(ctx, uuid)
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
