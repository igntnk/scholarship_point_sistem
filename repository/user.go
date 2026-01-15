package repository

import (
	"context"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	CreateUser(ctx context.Context, args db.CreateUserParams) (pgtype.UUID, error)
	GetApprovedUserByGradeBookNumber(ctx context.Context, gradeNumber string) (db.GetApprovedUserByGradeBookNumberRow, error)
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
