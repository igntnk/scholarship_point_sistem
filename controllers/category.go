package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/service"
	"github.com/igntnk/scholarship_point_system/service/models"
	"net/http"
	"strconv"
)

type categoryController struct {
	categoryService service.CategoryService
}

func NewCategoryController(
	categoryService service.CategoryService,
) Controller {
	return &categoryController{
		categoryService: categoryService,
	}
}

func (c categoryController) Register(r *gin.Engine) {
	group := r.Group("/category")
	group.POST("", c.Create)
	group.GET("", c.GetCategories)
	group.GET("/:uuid", c.GetByUUID)
	group.PUT("/:uuid", c.Update)
	group.DELETE("/:uuid", c.DeleteCategory)
}

func (c categoryController) Create(context *gin.Context) {

	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	createCategoryRequest := requests.CreateCategory{}
	if err = context.ShouldBindBodyWithJSON(&createCategoryRequest); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	var uuid string
	uuid, err = c.categoryService.CreateCategory(
		context,
		createCategoryRequest,
	)
	if err != nil {
		return
	}

	response := responses.CreateCategory{Uuid: uuid}
	context.JSON(http.StatusOK, createResponse(response))
	return
}

func (c categoryController) GetByUUID(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	var category models.Category
	category, err = c.categoryService.GetCategoryByUuid(context, context.Params.ByName("uuid"))
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(category))
}

func (c categoryController) GetCategories(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	queries := context.Request.URL.Query()
	strLimit := queries.Get("limit")
	if strLimit == "" {
		var categories []models.Category
		categories, err = c.categoryService.GetCategories(context)
		if err != nil {
			return
		}
		context.JSON(http.StatusOK, createResponse(categories))
		return
	}
	strOffset := queries.Get("offset")
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

	categories, totalRecords, err := c.categoryService.GetCategoriesWithPagination(context, limit, offset)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponseWithPagination(categories, limit, offset, totalRecords))
}

func (c categoryController) DeleteCategory(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	if err = c.categoryService.DeleteCategory(context, context.Params.ByName("uuid")); err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Запись успешно удалена"))
}

func (c categoryController) Update(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	updateCategoryRequest := requests.UpdateCategory{
		UUID: context.Params.ByName("uuid"),
	}
	if err = context.ShouldBindBodyWithJSON(&updateCategoryRequest); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	err = c.categoryService.UpdateCategory(context, updateCategoryRequest)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Запись успешно обновлена"))
}
