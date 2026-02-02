package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-viper/mapstructure/v2"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/igntnk/scholarship_point_system/service"
	"net/http"
	"strconv"
)

type constantController struct {
	m               middleware.Middleware
	constantService service.ConstantService
}

func NewConstantController(
	m middleware.Middleware,
	constantService service.ConstantService,
) Controller {
	return &constantController{
		m:               m,
		constantService: constantService,
	}
}

func (c *constantController) Register(r *gin.Engine) {
	g := r.Group("/constant", c.m.CheckAccess)

	g.GET("/grades_amount", c.GetGradesAmount)
	g.PUT("/grades_amount", c.UpdateGradesAmount)
}

func (c *constantController) GetGradesAmount(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	amount, err := c.constantService.GetGradeAmountsConstant(context)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(responses.Constant{Constant: strconv.Itoa(amount)}))
}

func (c *constantController) UpdateGradesAmount(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	reqMap := map[string]interface{}{}
	if err = context.ShouldBindBodyWithJSON(&reqMap); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	req := requests.Constant{}
	cfg := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &req,
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return
	}

	err = decoder.Decode(reqMap)
	if err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	constInt, err := strconv.Atoi(req.Constant)
	if err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	if err = c.constantService.UpdateGradeAmountsConstant(context, constInt); err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Константа успешно обновлена"))
}
