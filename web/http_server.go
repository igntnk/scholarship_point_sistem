package web

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/config"
	"github.com/igntnk/scholarship_point_system/controllers"
	"github.com/igntnk/scholarship_point_system/middleware"
	"github.com/rs/zerolog"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

type httpServer struct {
	Logger                zerolog.Logger
	Router                *gin.Engine
	srv                   http.Server
	checkPermissionRoutes map[string]struct{}
}

func New(logger zerolog.Logger, port int, corsCfg config.CorsConfig,
	ctrl ...controllers.Controller) (HttpServer, error) {

	r := gin.New()
	r.Use(middleware.NewCORS(corsCfg))
	r.Use(gin.Recovery())

	for i := 0; i < len(ctrl); i++ {
		ctrl[i].Register(r)
	}

	checkPermissionRoutes := map[string]struct{}{}

	for _, route := range r.Routes() {
		checkPermissionRoutes[unifyRelativePath(route.Path, route.Method)] = struct{}{}
	}

	return &httpServer{
		Router:                r,
		checkPermissionRoutes: checkPermissionRoutes,
		Logger:                logger.With().Str("Server", "HTTP").Logger(),
		srv: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
	}, nil
}

type HttpServer interface {
	ListenAndServe() error
	Routes() gin.RoutesInfo
	Shutdown(ctx context.Context) error
	GetRoutes() map[string]struct{}
}

func (h *httpServer) ListenAndServe() error {
	return h.srv.ListenAndServe()
}

func (h *httpServer) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

func (h *httpServer) Routes() gin.RoutesInfo {
	return h.Router.Routes()
}

func unifyRelativePath(path, method string) string {
	pathBuilder := strings.Builder{}
	pathBuilder.WriteString(method)
	pathBuilder.WriteString(" - ")

	if !strings.HasPrefix(path, "/") {
		pathBuilder.WriteString("/")
	}

	if strings.HasSuffix(path, "/") {
		pathBuilder.WriteString(path[:utf8.RuneCountInString(path)-1])
	} else {
		pathBuilder.WriteString(path)
	}

	resultPath := pathBuilder.String()

	re := regexp.MustCompile(`:[^/]+`)
	return re.ReplaceAllString(resultPath, ":var")
}

func (h *httpServer) GetRoutes() map[string]struct{} {
	return h.checkPermissionRoutes
}
