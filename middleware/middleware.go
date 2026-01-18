package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/igntnk/scholarship_point_system/errors/authorization"
	"github.com/igntnk/scholarship_point_system/jwk"
	"github.com/igntnk/scholarship_point_system/service"
	"net/http"
	"strings"
)

type Middleware interface {
	Authorize(c *gin.Context)
	CheckAccess(c *gin.Context)
}

type middleware struct {
	PermissionService service.PermissionService
	jwkey             jwk.JWKSigner
}

func NewMiddleware(PermissionService service.PermissionService, jwk jwk.JWKSigner) Middleware {
	return &middleware{
		PermissionService: PermissionService,
		jwkey:             jwk,
	}
}

func (m *middleware) Authorize(c *gin.Context) {
	m.authorize(c)
	c.Next()
	return
}

func (m *middleware) authorize(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	t := strings.Split(authHeader, " ")
	if len(t) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authorization.UnauthorizedErr.Error()})
	}

	authToken := t[1]

	claims, err := m.jwkey.ParseAccessToken(authToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authorization.TokenExpiredErr.Error()})
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authorization.UnauthorizedErr.Error()})
		return
	}

	if claims.ID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authorization.TokenDeniedErr.Error()})
		return
	}

	jwk.WithClaims(c, *claims)
	return
}

func (m *middleware) CheckAccess(c *gin.Context) {
	m.checkAccess(c, true)
}

func (m *middleware) checkAccess(c *gin.Context, recursive bool) {
	claims := jwk.ClaimsFromContext(c)
	if claims == nil {
		if recursive {
			m.authorize(c)
			m.checkAccess(c, false)
			return
		}

		if !c.IsAborted() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authorization.UnauthorizedErr.Error()})
		}

		return
	}

	if !claims.IsAdmin {
		path := c.Request.URL.Path
		hasAccess, err := m.PermissionService.CheckUserHasPermission(c, claims.User.UUID, path)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authorization.UnauthorizedErr.Error()})
			return
		}

		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": authorization.HasNoPermissionErr.Error()})
			return
		}
	}

	c.Next()
}
