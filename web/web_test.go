package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	test_suite "github.com/javitab/go-web/tests"
	"github.com/stretchr/testify/assert"
)

func AppRouter() *gin.Engine {
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Setup Security Headers
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Next()
	})

	// Return the router instance
	return router

}

func TestWebRouter(t *testing.T) {
	// Start Web Server
	router := test_suite.AppRouter()

	WebRouterGroup(router)

	// Send HTTP Request for Hello World
	expectedResponse :=
		`Web Test Successful`
	req, _ := http.NewRequest("GET", "/web/test", nil)
	req.Header.Set("content-type", "application/json;")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	assert.Equal(t, expectedResponse, string(responseData))
	assert.Equal(t, http.StatusOK, w.Code)
}
