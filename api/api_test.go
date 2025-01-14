package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	test_suite "github.com/javitab/go-web/tests"
	"github.com/stretchr/testify/assert"
)

func TestAPIRouter(t *testing.T) {
	// Start Web Server
	router := test_suite.AppRouter()

	// Setup route group for API
	ApiRouterGroup(router)

	// Send HTTP Request for Hello World
	expectedResponse :=
		`{"message":"Uniform API"}`
	req, _ := http.NewRequest("GET", "/api/", nil)
	req.Header.Set("content-type", "application/json;")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	assert.Equal(t, expectedResponse, string(responseData))
	assert.Equal(t, http.StatusOK, w.Code)
}
