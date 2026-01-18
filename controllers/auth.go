package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/errors/authorization"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/jwk"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/igntnk/scholarship_point_system/service"
	"net/http"
)

type AuthController struct {
	authService service.AuthService
	m           middleware.Middleware
}

func NewAuthController(
	authService service.AuthService,
	m middleware.Middleware,
) Controller {
	return &AuthController{
		authService: authService,
		m:           m,
	}
}

func (c *AuthController) Register(r *gin.Engine) {
	group := r.Group("/auth")
	group.POST("/change-password", c.m.Authorize, c.ChangePassword)
	group.POST("/signin", c.SignIn)
	group.POST("/signup", c.SignUp)
	group.POST("/refresh-token", c.RefreshToken)
}

func (c *AuthController) ChangePassword(context *gin.Context) {
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

	request := requests.ChangePassword{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, validation.WrongInputErr)
		return
	}

	if err = c.authService.ChangePassword(
		context,
		accessClaims.(jwk.SPSAccessClaims).User.UUID,
		request.Password,
	); err != nil {
		return
	}

	context.JSON(http.StatusOK, createResponse("Пароль успешно изменен"))
}

func (c *AuthController) SignIn(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.SingIn{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, validation.WrongInputErr)
		return
	}

	accessToken, refreshToken, err := c.authService.SignIn(context, request.Email, request.Password)
	if err != nil {
		return
	}

	response := responses.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	context.JSON(http.StatusOK, createResponse(response))
}

func (c *AuthController) SignUp(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	request := requests.CreateUser{}
	if err = context.ShouldBindBodyWithJSON(&request); err != nil {
		err = errors.Join(err, validation.WrongInputErr)
		return
	}

	_, accessToken, refreshToken, err := c.authService.SignUp(context, request)
	if err != nil {
		return
	}

	response := responses.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	context.JSON(http.StatusOK, createResponse(response))
}

func (c *AuthController) RefreshToken(context *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			processHttpError(context, err)
		}
	}()

	authHeader := context.Request.Header.Get("Authorization")
	accessToken, refreshToken, err := c.authService.RefreshToken(context, authHeader)
	if err != nil {
		return
	}

	response := responses.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	context.JSON(http.StatusOK, createResponse(response))
}
