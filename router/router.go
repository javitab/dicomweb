package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/javitab/go-web/api"
	"github.com/javitab/go-web/auth"
	docs "github.com/javitab/go-web/docs"
	"github.com/javitab/go-web/static_web"
	"github.com/javitab/go-web/web"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AppRouter() *gin.Engine {
	// Set the router as the default one shipped with Gin
	router := gin.Default()
	// expectedHost := os.Getenv("HTTP_HOST")

	// Setup CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "Accept", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup Security Headers
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// c.Header("Referrer-Policy", "strict-origin")
		// c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Next()
	})

	docs.SwaggerInfo.Title = "Go Web"
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve frontend static files
	web_fs, _ := static_web.HTTPFS()
	router.StaticFS("/static", web_fs)

	// Setup route group for API
	api.ApiRouterGroup(router)

	// Setup route group for web
	web.WebRouterGroup(router)

	// Setup route group for auth
	auth.AuthRouterGroup(router)

	// Return the router instance
	return router
}
