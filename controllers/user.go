package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/service"
	"net/http"
	"strconv"
)

type userController struct {
	userService service.UserService
}

func NewUserController(
	userService service.UserService,
) Controller {
	return &userController{
		userService: userService,
	}
}

func (c userController) Register(r *gin.Engine) {
	group := r.Group("/user")
	group.GET("/simple", c.GetSimpleUserList)
	group.GET("/simple/:uuid", c.GetSimpleUserByUUID)
	group.POST("")
	group.PUT("/:uuid", c.UpdateUser)
}

func (c userController) GetSimpleUserList(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	queries := context.Request.URL.Query()
	strLimit := queries.Get("limit")
	if strLimit == "" {
		users, err := c.userService.GetSimpleUserList(context)
		if err != nil {
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"data": users,
		})
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

	users, totalRecords, err := c.userService.GetSimpleUserListWithPagination(context, limit, offset)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"total_records": totalRecords,
			"limit":         limit,
			"offset":        offset,
			"selected":      len(users),
		},
	})
}

func (c userController) GetSimpleUserByUUID(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid := context.Param("uuid")

	user, err := c.userService.GetSimpleUserByUUID(context, uuid)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": user})
}

func (c userController) CreateUser(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.CreateUser{}
	if err = context.ShouldBindJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	uuid, err := c.userService.CreateUser(context, request)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": uuid})
}

func (c userController) UpdateUser(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid := context.Param("uuid")
	request := requests.UpdateUser{
		UUID: uuid,
	}
	if err = context.ShouldBindJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	if err = c.userService.UpdateUser(context, request); err != nil {
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": "Информация о пользователе успешно обновлена"})
}
