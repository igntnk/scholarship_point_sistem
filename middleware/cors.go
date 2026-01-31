package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/config"
	"time"
)

func NewCORS(cfg config.CorsConfig) gin.HandlerFunc {
	c := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           12 * time.Hour,
	}

	if cfg.AllowAll || len(cfg.AllowedOrigins) == 0 {
		c.AllowAllOrigins = true
	} else {
		c.AllowOrigins = cfg.AllowedOrigins
	}

	return cors.New(c)
}
