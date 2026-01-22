package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/errors/authorization"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/jwk"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/igntnk/scholarship_point_system/service"
	"github.com/igntnk/scholarship_point_system/service/models"
	"net/http"
	"strconv"
)

type achievementController struct {
	achievementService service.AchievementService
	m                  middleware.Middleware
}

func NewAchievementController(
	achievementService service.AchievementService,
	m middleware.Middleware,
) Controller {
	return &achievementController{
		achievementService: achievementService,
		m:                  m,
	}
}

func (c *achievementController) Register(r *gin.Engine) {
	group := r.Group("/achievement", c.m.CheckAccess)
	group.GET("/by_token", c.m.CheckAccess, c.ListMyAchievements)
	group.GET("/by_user_uuid/:uuid", c.m.CheckAccess)
	group.POST("", c.m.CheckAccess, c.CreateAchievement)
	group.DELETE("", c.m.CheckAccess, c.DeleteAchievement)
	group.PUT("", c.m.CheckAccess, c.UpdateAchievement)
}

func (c *achievementController) ListMyAchievements(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	accessClaims, ok := context.Get(jwk.ClaimsContextKey)
	if !ok {
		err = authorization.UnauthorizedErr
		return
	}

	queryParams := context.Request.URL.Query()

	strLimit := queryParams.Get("limit")
	strOffset := queryParams.Get("offset")

	if strLimit == "" {
		modelAchievements, err := c.achievementService.GetUserAchievements(context, accessClaims.(jwk.SPSAccessClaims).User.UUID)
		if err != nil {
			return
		}

		respAchievement := make([]models.SimpleAchievement, len(modelAchievements))
		for i, modelAchievement := range modelAchievements {
			respAchievement[i] = models.SimpleAchievement{
				UUID:           modelAchievement.UUID,
				AttachmentLink: modelAchievement.AttachmentLink,
				Status:         modelAchievement.Status,
				Comment:        modelAchievement.Comment,
				CategoryName:   modelAchievement.CategoryName,
				PointAmount:    modelAchievement.PointAmount,
			}
		}

		context.JSON(http.StatusOK, createResponse(respAchievement))
		return
	}

	if strOffset == "" {
		strOffset = "0"
	}

	limit, err := strconv.Atoi(strLimit)
	if err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}
	offset, err := strconv.Atoi(strOffset)
	if err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	achievements, totalRecords, err := c.achievementService.GetUserAchievementsWithPagination(
		context,
		accessClaims.(jwk.SPSAccessClaims).User.UUID,
		limit,
		offset,
	)
	if err != nil {
		return
	}

	respAchievement := make([]models.SimpleAchievement, len(achievements))
	for i, modelAchievement := range achievements {
		respAchievement[i] = models.SimpleAchievement{
			UUID:           modelAchievement.UUID,
			AttachmentLink: modelAchievement.AttachmentLink,
			Status:         modelAchievement.Status,
			Comment:        modelAchievement.Comment,
			CategoryName:   modelAchievement.CategoryName,
			PointAmount:    modelAchievement.PointAmount,
		}
	}

	context.JSON(http.StatusOK, createResponseWithPagination(respAchievement, limit, offset, totalRecords))
}

func (c *achievementController) ListUserAchievements(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	userUUID := context.Param("uuid")
	if userUUID == "" {
		err = errors.Join(validation.WrongInputErr, errors.New("Не найден uuid пользователя"))
		return
	}

	queryParams := context.Request.URL.Query()

	strLimit := queryParams.Get("limit")
	strOffset := queryParams.Get("offset")

	if strLimit == "" {
		modelAchievements, err := c.achievementService.GetUserAchievements(context, userUUID)
		if err != nil {
			return
		}

		respAchievement := make([]models.SimpleAchievement, len(modelAchievements))
		for i, modelAchievement := range modelAchievements {
			respAchievement[i] = models.SimpleAchievement{
				UUID:           modelAchievement.UUID,
				AttachmentLink: modelAchievement.AttachmentLink,
				Status:         modelAchievement.Status,
				Comment:        modelAchievement.Comment,
				CategoryName:   modelAchievement.CategoryName,
				PointAmount:    modelAchievement.PointAmount,
			}
		}

		context.JSON(http.StatusOK, createResponse(respAchievement))
		return
	}

	if strOffset == "" {
		strOffset = "0"
	}

	limit, err := strconv.Atoi(strLimit)
	if err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}
	offset, err := strconv.Atoi(strOffset)
	if err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	achievements, totalRecords, err := c.achievementService.GetUserAchievementsWithPagination(
		context,
		userUUID,
		limit,
		offset,
	)
	if err != nil {
		return
	}

	respAchievement := make([]models.SimpleAchievement, len(achievements))
	for i, modelAchievement := range achievements {
		respAchievement[i] = models.SimpleAchievement{
			UUID:           modelAchievement.UUID,
			AttachmentLink: modelAchievement.AttachmentLink,
			Status:         modelAchievement.Status,
			Comment:        modelAchievement.Comment,
			CategoryName:   modelAchievement.CategoryName,
			PointAmount:    modelAchievement.PointAmount,
		}
	}

	context.JSON(http.StatusOK, createResponseWithPagination(respAchievement, limit, offset, totalRecords))
}

func (c *achievementController) CreateAchievement(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()
	req := requests.UpsertAchievement{}
	if err := context.ShouldBindBodyWithJSON(&req); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	achievement := models.Achievement{
		AttachmentLink: req.AttachmentLink,
		Comment:        req.Comment,
	}

	categories := make([]models.Category, len(req.CategoryUUIDs))
	for i, uuid := range req.CategoryUUIDs {
		categories[i].UUID = uuid
	}
	achievement.Categories = categories

	uuid, err := c.achievementService.CreateAchievement(context, achievement)
	if err != nil {
		return
	}

	context.JSON(http.StatusCreated, createResponse(gin.H{
		"uuid": uuid,
	}))
}

func (c *achievementController) DeleteAchievement(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	achievementUUID := context.Param("uuid")

	err = c.achievementService.RemoveAchievement(context, achievementUUID)
	if err != nil {
		return
	}

	context.JSON(http.StatusCreated, createResponse("Запись успешно удалена"))
}

func (c *achievementController) UpdateAchievement(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()
	req := requests.UpsertAchievement{}
	if err := context.ShouldBindBodyWithJSON(&req); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	achievement := models.Achievement{
		AttachmentLink: req.AttachmentLink,
		Comment:        req.Comment,
	}

	categories := make([]models.Category, len(req.CategoryUUIDs))
	for i, uuid := range req.CategoryUUIDs {
		categories[i].UUID = uuid
	}
	achievement.Categories = categories

	err = c.achievementService.UpdateAchievement(context, achievement)
	if err != nil {
		return
	}

	context.JSON(http.StatusCreated, createResponse("Запись успешно обновлена"))
}
