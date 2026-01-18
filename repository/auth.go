package repository

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
)

type AuthRepository interface {
	ChangePassword(ctx context.Context, uuid, salt, password string) error
}

type authRepository struct {
	queries *db.Queries
}

func NewAuthRepository(pool db.DBTX) AuthRepository {
	return &authRepository{
		queries: db.New(pool),
	}
}

func (r *authRepository) ChangePassword(ctx context.Context, uuid, salt, password string) error {

	pgPassword, err := ParseToPgText(password)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	pgUuid, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	pgSalt, err := ParseToPgText(salt)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	args := db.ChangePasswordParams{
		Password: pgPassword,
		Uuid:     pgUuid,
		Salt:     pgSalt,
	}
	if err = r.queries.ChangePassword(ctx, args); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *authRepository) SignIn(ctx context.Context, email, password string) (string, string, error) {
	return "", "", nil
}
