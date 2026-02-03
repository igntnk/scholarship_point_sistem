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
	"net/http"
	"strconv"
)

type userController struct {
	userService service.UserService
	m           middleware.Middleware
}

func NewUserController(
	userService service.UserService,
	m middleware.Middleware,
) Controller {
	return &userController{
		userService: userService,
		m:           m,
	}
}

func (c userController) Register(r *gin.Engine) {
	group := r.Group("/user", c.m.CheckAccess)
	group.GET("/simple", c.GetSimpleUserList)
	group.GET("/simple/:uuid", c.GetSimpleUserByUUID)
	group.POST("", c.CreateUser)
	group.PUT("/:uuid", c.UpdateUser)
	group.PUT("/approve/:uuid", c.ApproveUser)
	group.PUT("/decline/:uuid", c.DeclineUser)

	self := r.Group("/user", c.m.Authorize)
	self.GET("/me", c.GetMe)
	self.PUT("/me", c.UpdateMe)
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

		context.JSON(http.StatusOK, createResponse(users))
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

	context.JSON(http.StatusOK, createResponseWithPagination(users, limit, offset, totalRecords))
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

	context.JSON(http.StatusOK, createResponse(user))
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

	context.JSON(http.StatusOK, createResponse(uuid))
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

	context.JSON(http.StatusOK, createResponse("Информация о пользователе успешно обновлена"))
}

func (c userController) GetMe(context *gin.Context) {
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

	userUUID := accessClaims.(jwk.SPSAccessClaims).User.UUID

	user, err := c.userService.GetSimpleUserByUUID(context, userUUID)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(user))
}

func (c userController) UpdateMe(context *gin.Context) {
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

	userUUID := accessClaims.(jwk.SPSAccessClaims).User.UUID

	request := requests.UpdateUser{
		UUID: userUUID,
	}
	if err = context.ShouldBindJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	if err = c.userService.UpdateUser(context, request); err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Информация о пользователе успешно обновлена"))
}

func (c userController) ApproveUser(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid := context.Param("uuid")
	if err = c.userService.ApproveUser(context, uuid); err != nil {
		return
	}
	context.JSON(http.StatusOK, createResponse("Пользователь успешно подтвержден"))
}

func (c userController) DeclineUser(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid := context.Param("uuid")

	if err = c.userService.DeclineUser(context, uuid); err != nil {
		return
	}
	context.JSON(http.StatusOK, createResponse("Пользователь успешно отклонен"))
}
