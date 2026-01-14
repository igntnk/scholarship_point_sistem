package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/service"
	"net/http"
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
}

func (c categoryController) Create(context *gin.Context) {

	var err error

	defer func() {
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}()

	createCategoryRequest := requests.CreateCategory{}
	if err = context.ShouldBindBodyWithJSON(&createCategoryRequest); err != nil {
		err = errors.Join(err, ErrParsingBody)
		return
	}

	var uuid string
	uuid, err = c.categoryService.CreateCategory(
		context,
		createCategoryRequest.Name,
		createCategoryRequest.Comment,
		createCategoryRequest.ParentUuid,
		createCategoryRequest.PointAmount,
	)
	if err != nil {
		return
	}

	response := responses.CreateCategory{Uuid: uuid}
	context.JSON(http.StatusOK, response)
	return
}
