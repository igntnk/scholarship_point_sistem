package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/igntnk/scholarship_point_system/service"
	"net/http"
	"strconv"
)

type permissionController struct {
	permissionService service.PermissionService
	m                 middleware.Middleware
}

func (c *permissionController) Register(r *gin.Engine) {
	g := r.Group("/permission", c.m.CheckAccess)
	roleG := g.Group("/role")
	groupG := g.Group("/group")

	roleG.GET("", c.GetRoleList)
	roleG.GET("/:uuid", c.GetRoleByUUID)
	roleG.DELETE("/:uuid", c.DeleteRole)
	roleG.PUT("/:uuid", c.UpdateRole)
	roleG.POST("", c.CreateRole)

	groupG.GET("", c.GetGroupList)
	groupG.GET("/:uuid", c.GetGroupByUUID)
	groupG.DELETE("/:uuid", c.DeleteGroup)
	groupG.PUT("/:uuid", c.UpdateGroup)
	groupG.POST("", c.CreateGroup)
}

func NewPermissionController(
	permissionService service.PermissionService,
	m middleware.Middleware,
) Controller {
	return &permissionController{
		permissionService: permissionService,
		m:                 m,
	}
}

func (c *permissionController) GetRoleByUUID(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid, ok := context.Params.Get("uuid")
	if !ok {
		err = parsing.InputDataErr
		return
	}

	role, err := c.permissionService.GetRoleByUUID(context, uuid)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(role))
}

func (c *permissionController) GetGroupByUUID(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid, ok := context.Params.Get("uuid")
	if !ok {
		err = parsing.InputDataErr
		return
	}

	group, err := c.permissionService.GetGroupByUUID(context, uuid)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(group))
}

func (c *permissionController) GetRoleList(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	query := context.Request.URL.Query()
	strLimit := query.Get("limit")
	strOffset := query.Get("offset")
	if strLimit == "" {
		roles, err := c.permissionService.GetRoleList(context)
		if err != nil {
			return
		}

		context.JSON(http.StatusOK, createResponse(roles))
		return
	}

	limit, err := strconv.Atoi(strLimit)
	if err != nil {
		err = parsing.InputDataErr
		return
	}

	offset, err := strconv.Atoi(strOffset)
	if err != nil {
		err = parsing.InputDataErr
		return
	}

	roles, totalRecords, err := c.permissionService.GetRoleListWithPagination(context, limit, offset)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponseWithPagination(roles, limit, offset, totalRecords))
}

func (c *permissionController) GetGroupList(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	query := context.Request.URL.Query()
	strLimit := query.Get("limit")
	strOffset := query.Get("offset")
	if strLimit == "" {
		groups, err := c.permissionService.GetGroupList(context)
		if err != nil {
			return
		}

		context.JSON(http.StatusOK, createResponse(groups))
		return
	}

	limit, err := strconv.Atoi(strLimit)
	if err != nil {
		err = parsing.InputDataErr
		return
	}

	offset, err := strconv.Atoi(strOffset)
	if err != nil {
		err = parsing.InputDataErr
		return
	}

	groups, totalRecords, err := c.permissionService.GetGroupListWithPagination(context, limit, offset)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponseWithPagination(groups, limit, offset, totalRecords))
}

func (c *permissionController) DeleteRole(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid, ok := context.Params.Get("uuid")
	if !ok {
		err = parsing.InputDataErr
		return
	}

	err = c.permissionService.DeleteRole(context, uuid)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Роль успешно удалена"))
}

func (c *permissionController) DeleteGroup(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	uuid, ok := context.Params.Get("uuid")
	if !ok {
		err = parsing.InputDataErr
		return
	}

	err = c.permissionService.DeleteGroup(context, uuid)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Группа успешно удалена"))
}

func (c *permissionController) UpdateRole(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.Role{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	if err = c.permissionService.UpdateRoleWithMembers(context, request); err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Роль успешно обновлена"))
}

func (c *permissionController) UpdateGroup(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.Group{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	if err = c.permissionService.UpdateGroupWithRolesAndResources(context, request); err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Группа успешно обновлена"))
}

func (c *permissionController) CreateRole(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.Role{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	uuid, err := c.permissionService.CreateRole(context, request)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(uuid))
}

func (c *permissionController) CreateGroup(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.Group{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, parsing.InputDataErr)
		return
	}

	uuid, err := c.permissionService.CreateGroup(context, request)
	if err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse(uuid))
}
