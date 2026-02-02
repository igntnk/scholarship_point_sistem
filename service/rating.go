package service

import (
	"context"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/repository"
	"strings"
)

type RatingService interface {
	GetRating(context.Context, requests.GetRating) (users []responses.User, totalRecords int, err error)
}

type ratingService struct {
	userRepo repository.UserRepository
}

func NewRatingService(
	userRepo repository.UserRepository,
) RatingService {
	return &ratingService{
		userRepo: userRepo,
	}
}

func (s *ratingService) GetRating(
	ctx context.Context,
	req requests.GetRating,
) (
	users []responses.User,
	totalRecords int,
	err error,
) {
	searchWords := strings.Split(req.SearchString, " ")

	modUsers, totalRecords, err := s.userRepo.GetRating(
		ctx,
		searchWords,
		req.Valid,
		req.Winners,
		req.Limit,
		req.Offset,
	)
	if err != nil {
		return nil, 0, err
	}

	return modUsers, totalRecords, nil
}
