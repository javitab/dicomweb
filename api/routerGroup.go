package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javitab/go-web/middlewares"
)

func ApiRouterGroup(router *gin.Engine) *gin.RouterGroup {
	api := router.Group("/api", middlewares.CheckAuth)
	{
		apiHandler := func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Uniform API",
			})
		}
		api.GET("", apiHandler)
		api.GET("/", apiHandler)
		api.Any("/server_events", ServerEventHandler)
	}
	return api
}
