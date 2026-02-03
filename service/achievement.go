package service

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
)

type AchievementService interface {
	GetUserAchievements(ctx context.Context, uuid string) ([]responses.SimpleAchievement, error)
	GetUserAchievementsWithPagination(ctx context.Context, uuid string, limit, offset int) ([]responses.SimpleAchievement, int, error)
	GetAchievementByUUID(ctx context.Context, uuid string) (responses.FullAchievement, error)
	CreateAchievement(ctx context.Context, userUUID string, achievement requests.UpsertAchievement) (string, error)
	UpdateAchievement(ctx context.Context, achievement requests.UpsertAchievement) error
	RemoveAchievement(ctx context.Context, uuid string) error
	ApproveAchievement(ctx context.Context, uuid string) error
	DeclineAchievement(ctx context.Context, uuid string) error
}

type achievementService struct {
	achievementRepo repository.AchievementRepository
	userRepo        repository.UserRepository
}

func NewAchievementService(
	r repository.AchievementRepository,
	u repository.UserRepository,
) AchievementService {
	return &achievementService{
		achievementRepo: r,
		userRepo:        u,
	}
}

func (s *achievementService) GetUserAchievements(ctx context.Context, uuid string) ([]responses.SimpleAchievement, error) {
	_, err := s.userRepo.GetSimpleUserByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, validation.NoDataFoundErr) {
			return nil, errors.Join(err, errors.New("Такого пользователя нет"))
		}
		return nil, err
	}

	modelAchievements, err := s.achievementRepo.GetUserAchievements(ctx, uuid)
	if err != nil {
		return nil, err
	}

	response := make([]responses.SimpleAchievement, len(modelAchievements))
	for i, achievement := range modelAchievements {
		response[i] = responses.SimpleAchievement{
			UUID:           achievement.UUID,
			Comment:        achievement.Comment,
			Status:         achievement.Status,
			CategoryName:   achievement.CategoryName,
			CategoryUUID:   achievement.CategoryUUID,
			PointAmount:    achievement.PointAmount,
			AttachmentLink: achievement.AttachmentLink,
		}
	}

	return response, nil
}

func (s *achievementService) GetUserAchievementsWithPagination(ctx context.Context, uuid string, limit, offset int) ([]responses.SimpleAchievement, int, error) {
	_, err := s.userRepo.GetSimpleUserByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, validation.NoDataFoundErr) {
			return nil, 0, errors.Join(err, errors.New("Такого пользователя нет"))
		}
		return nil, 0, err
	}

	modelAchievements, totalRecords, err := s.achievementRepo.GetUserAchievementsWithPagination(ctx, uuid, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	response := make([]responses.SimpleAchievement, len(modelAchievements))
	for i, achievement := range modelAchievements {
		response[i] = responses.SimpleAchievement{
			UUID:           achievement.UUID,
			Comment:        achievement.Comment,
			Status:         achievement.Status,
			CategoryName:   achievement.CategoryName,
			PointAmount:    achievement.PointAmount,
			AttachmentLink: achievement.AttachmentLink,
		}
	}

	return response, totalRecords, nil
}

func (s *achievementService) GetAchievementByUUID(ctx context.Context, uuid string) (responses.FullAchievement, error) {
	return s.achievementRepo.GetAchievementByUUID(ctx, uuid)
}

func (s *achievementService) CreateAchievement(ctx context.Context, userUUID string, a requests.UpsertAchievement) (string, error) {
	if a.AttachmentLink == "" {
		return "", errors.Join(validation.WrongInputErr, errors.New("Пустая ссылка для вложения"))
	}

	return s.achievementRepo.CreateAchievement(ctx, userUUID, a)
}

func (s *achievementService) UpdateAchievement(ctx context.Context, a requests.UpsertAchievement) error {
	if a.AttachmentLink == "" {
		return errors.Join(validation.WrongInputErr, errors.New("Пустая ссылка для вложения"))
	}

	if len(a.CategoryUUID) == 0 {
		return errors.Join(validation.WrongInputErr, errors.New("Нет категории у достижения"))
	}

	dbAchievement, err := s.achievementRepo.GetSimpleUserAchievementByUUID(ctx, a.UUID)
	if err != nil {
		return err
	}

	dbCats, err := s.achievementRepo.GetAchievementCategories(ctx, dbAchievement.UUID)
	if err != nil {
		return err
	}

	hasDiff := false
	for _, cat := range a.Subcategories {
		found := false
		for _, dbCat := range dbCats {
			if dbCat.UUID == cat.UUID {
				found = true
				break
			}
		}
		if !found {
			found = cat.UUID == a.CategoryUUID
		}
		if !found {
			hasDiff = true
			break
		}
	}

	if !hasDiff && a.AttachmentLink == dbAchievement.AttachmentLink {
		return s.achievementRepo.UpdateAchievementDescFields(ctx, models.SimpleAchievement{
			UUID:           a.UUID,
			AttachmentLink: a.AttachmentLink,
			Comment:        a.Comment,
		})
	}

	return s.achievementRepo.UpdateAchievementFull(ctx, a)
}

func (s *achievementService) RemoveAchievement(ctx context.Context, uuid string) error {
	_, err := s.achievementRepo.GetSimpleUserAchievementByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	return s.achievementRepo.MakeAchievementRemoved(ctx, uuid)
}

func (s *achievementService) ApproveAchievement(ctx context.Context, uuid string) error {
	_, err := s.achievementRepo.GetSimpleUserAchievementByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	return s.achievementRepo.MakeAchievementApproved(ctx, uuid)
}

func (s *achievementService) DeclineAchievement(ctx context.Context, uuid string) error {
	_, err := s.achievementRepo.GetSimpleUserAchievementByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	return s.achievementRepo.MakeAchievementDeclined(ctx, uuid)
}
