package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	test_suite "github.com/javitab/go-web/tests"
	"github.com/stretchr/testify/assert"
)

func TestAuthRouter(t *testing.T) {
	tearDown := test_suite.SetupSuite(t)
	defer tearDown(t)

	// Start Web Server
	router := test_suite.AppRouter()

	AuthRouterGroup(router)

	// Send HTTP Request for Hello World
	expectedResponse :=
		`{"message":"Auth router","path":"/auth"}`
	req, _ := http.NewRequest("GET", "/auth", nil)
	req.Header.Set("content-type", "application/json;")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	assert.Equal(t, expectedResponse, string(responseData))
	assert.Equal(t, http.StatusOK, w.Code)
}
