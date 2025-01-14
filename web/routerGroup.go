package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WebRouterGroup(router *gin.Engine) *gin.RouterGroup {
	web := router.Group("/web")
	{
		web.GET("/hello", helloHandler)
		web.GET("/test", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/plain", []byte("Web Test Successful"))
		})
	}
	return web
}
