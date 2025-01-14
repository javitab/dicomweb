package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javitab/go-web/middlewares"
)

func AuthRouterGroup(router *gin.Engine) *gin.RouterGroup {
	auth := router.Group("/auth")
	{
		authHandler := func(c *gin.Context) {

			if c.Request.Method == "GET" {
				c.JSON(http.StatusOK, gin.H{
					"message": "Auth router",
					"path":    c.Request.URL.Path,
				})
			} else if c.Request.Method == "POST" {
				CreateUser(c)
			} else {
				c.JSON(http.StatusMethodNotAllowed, gin.H{
					"error": "Method not allowed",
				})
			}

		}
		auth.GET("", authHandler)
		auth.GET("/", authHandler)
		auth.POST("/create_user", CreateUser)
		auth.POST("/login", LoginUser)
		auth.POST("/generate_jwt", GetJWTFromAPIKey)
		auth.POST("/update_user", middlewares.CheckAuth, UpdateUser)
		auth.GET("/user", middlewares.CheckAuth, GetUser)
		auth.GET("/group", middlewares.CheckAuth, GetGroup)
		auth.GET("/sec_point", middlewares.CheckAuth, GetSecPoint)
		auth.POST("/generate_api_key", middlewares.CheckAuth, GenerateAPIKey)

		return auth
	}
}
