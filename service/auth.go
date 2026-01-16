package service

import (
	"context"
	"github.com/igntnk/scholarship_point_system/repository"
)

type AuthService interface {
	ChangePassword(ctx context.Context, uuid, password string) error
}

type authService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
}

func NewAuthService(
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
) AuthService {
	return &authService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (s *authService) ChangePassword(ctx context.Context, uuid, password string) error {
	_, err := s.userRepo.GetSimpleUserByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
