package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/errors/authorization"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/jwk"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/igntnk/scholarship_point_system/service"
	"io"
	"net/http"
)

type ratingController struct {
	m             middleware.Middleware
	ratingService service.RatingService
}

func NewRatingController(
	m middleware.Middleware,
	ratingService service.RatingService,
) Controller {
	return &ratingController{
		m:             m,
		ratingService: ratingService,
	}
}

func (c *ratingController) Register(r *gin.Engine) {
	g := r.Group("/rating", c.m.CheckAccess)

	g.POST("", c.GetRating)
	g.GET("/short_info", c.GetShortInfo)
}

func (c *ratingController) GetRating(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	req := requests.GetRating{}
	if err = context.ShouldBindBodyWithJSON(&req); err != nil {
		if !errors.As(err, &io.EOF) {
			err = errors.Join(err, parsing.InputDataErr)
			return
		}
	}

	resp, totalRecs, err := c.ratingService.GetRating(context, req)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponseWithPagination(resp, req.Limit, req.Offset, totalRecs))
}

func (c *ratingController) GetShortInfo(context *gin.Context) {
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

	result, err := c.ratingService.GetShortInfo(context, accessClaims.(jwk.SPSAccessClaims).User.UUID)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(result))
}
