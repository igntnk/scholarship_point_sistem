package service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/errors/authorization"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/jwk"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"
)

type AuthService interface {
	ActualizeAdmin(ctx context.Context, email, password string) (string, error)
	ChangePassword(ctx context.Context, uuid, password string) error
	SignIn(ctx context.Context, email, password string) (string, string, error)
	SignUp(ctx context.Context, user requests.CreateUser) (string, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}

type authService struct {
	authRepo             repository.AuthRepository
	userRepo             repository.UserRepository
	permissionRepo       repository.PermissionRepository
	passwordManager      PasswordManager
	jwkey                jwk.JWKSigner
	AccessTokenDuration  int
	RefreshTokenDuration int
	AdminGroupName       string
}

func NewAuthService(
	authRepo repository.AuthRepository,
	userRepo repository.UserRepository,
	passwordManager PasswordManager,
	permissionRepo repository.PermissionRepository,
	jwkey jwk.JWKSigner,
	AccessTokenDuration int,
	RefreshTokenDuration int,
	AdminGroupName string,
) AuthService {
	return &authService{
		authRepo:             authRepo,
		userRepo:             userRepo,
		passwordManager:      passwordManager,
		permissionRepo:       permissionRepo,
		jwkey:                jwkey,
		AccessTokenDuration:  AccessTokenDuration,
		RefreshTokenDuration: RefreshTokenDuration,
		AdminGroupName:       AdminGroupName,
	}
}

func (s *authService) ActualizeAdmin(ctx context.Context, email, password string) (string, error) {
	var userUUID string
	user, err := s.userRepo.GetUserWithCredentialsByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, validation.NoDataFoundErr) {
			return "", err
		}

		userUUID, _, _, err = s.SignUp(ctx, requests.CreateUser{
			Name:            "Администратор",
			SecondName:      "Главный",
			GradebookNumber: "0-0",
			Email:           email,
			Password:        password,
		})
		if err != nil {
			return "", err
		}
	} else {
		userUUID = user.UUID
	}

	err = s.passwordManager.CompareHashAndPassword(user.HashedPassword, user.Salt, password)
	if err == nil {
		return user.UUID, nil
	}

	if err = s.ChangePassword(ctx, userUUID, password); err != nil {
		return "", err
	}

	return user.UUID, nil
}

func (s *authService) ChangePassword(ctx context.Context, uuid, password string) error {
	user, err := s.userRepo.GetSimpleUserByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	if err = s.passwordManager.ValidatePassword(password); err != nil {
		return err
	}

	hashedPassword, salt, err := s.passwordManager.HashPassword(password)
	if err != nil {
		return err
	}

	return s.authRepo.ChangePassword(ctx, user.UUID, salt, hashedPassword)
}

func (s *authService) SignIn(ctx context.Context, email, password string) (string, string, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if len(email) == 0 {
		return "", "", authorization.HasNoEmailErr
	}

	if len(password) == 0 {
		return "", "", authorization.HasNoPasswordErr
	}

	user, err := s.userRepo.GetUserWithCredentialsByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	err = s.passwordManager.CompareHashAndPassword(user.HashedPassword, user.Salt, password)
	if err != nil {
		return "", "", err
	}

	groups, err := s.permissionRepo.GetUserGroups(ctx, user.UUID)
	if err != nil {
		return "", "", err
	}

	return s.createAccessAndRefreshToken(models.SimpleUser{
		UUID:       user.UUID,
		Name:       user.Name,
		SecondName: user.SecondName,
		Patronymic: user.Patronymic,
		Email:      user.Email,
	}, groups)

}

func (s *authService) SignUp(ctx context.Context, user requests.CreateUser) (string, string, string, error) {
	_, err := s.userRepo.GetApprovedUserByGradeBookNumber(ctx, user.GradebookNumber)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", errors.Join(err, unexpected.RequestErr)
		}
	} else {
		return "", "", "", validation.RecordAlreadyExistsErr
	}

	if user.Name == "" {
		return "", "", "", errors.Join(validation.WrongInputErr, errors.New("Отсуствует имя пользователя"))
	}

	if user.GradebookNumber == "" {
		return "", "", "", errors.Join(validation.WrongInputErr, errors.New("Отсуствует номер зачетной книжки"))
	}

	if user.Email == "" {
		return "", "", "", errors.Join(validation.WrongInputErr, errors.New("Отсутсвует почта"))
	}

	if err = s.passwordManager.ValidatePassword(user.Password); err != nil {
		return "", "", "", err
	}

	hashedPassword, salt, err := s.passwordManager.HashPassword(user.Password)
	if err != nil {
		return "", "", "", err
	}

	userUUID, err := s.userRepo.CreateUser(ctx, models.UserWithCredentials{
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
	if err != nil {
		return "", "", "", err
	}

	access, refresh, err := s.createAccessAndRefreshToken(models.SimpleUser{
		UUID:       userUUID,
		Name:       user.Name,
		SecondName: user.SecondName,
		Patronymic: user.Patronymic,
		Email:      user.Email,
	}, nil)
	if err != nil {
		return "", "", "", err
	}

	return userUUID, access, refresh, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	refreshToken = strings.TrimSpace(refreshToken)

	refreshToken = strings.TrimLeft(refreshToken, "Bearer ")
	if len(refreshToken) == 0 {
		return "", "", errors.Join(authorization.TokenDeniedErr, errors.New("Токен обновления не найден"))
	}

	refreshClaims, err := s.jwkey.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	simpleUser, err := s.userRepo.GetSimpleUserByUUID(ctx, refreshClaims.UserUUID)
	if err != nil {
		return "", "", err
	}

	groups, err := s.permissionRepo.GetUserGroups(ctx, simpleUser.UUID)
	if err != nil {
		return "", "", err
	}

	return s.createAccessAndRefreshToken(simpleUser, groups)
}

func (s *authService) createAccessAndRefreshToken(
	user models.SimpleUser,
	groups []models.SimpleGroup,
) (string, string, error) {
	tokenID, err := uuid.NewUUID()
	if err != nil {
		return "", "", errors.Join(err, unexpected.InternalErr)
	}

	isAdmin := false
	for _, group := range groups {
		if group.Name == s.AdminGroupName {
			isAdmin = true
			break
		}
	}

	accessClaims := jwk.SPSAccessClaims{
		User: jwk.User{
			UUID:       user.UUID,
			Name:       user.Name,
			SecondName: user.SecondName,
			Patronymic: user.Patronymic,
			LastLogin:  time.Now().UTC().Unix(),
		},
		Data:    nil,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "SPS",
			Subject:   user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(s.AccessTokenDuration))),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        tokenID.String(),
		},
	}
	accessToken, err := s.jwkey.SignToken(accessClaims)
	if err != nil {
		return "", "", err
	}

	refClaims := jwk.SPSRefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "SPS",
			Subject:   user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(s.RefreshTokenDuration))),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        tokenID.String(),
		},
		Data:     nil,
		UserUUID: user.UUID,
	}

	refreshToken, err := s.jwkey.SignToken(refClaims)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
