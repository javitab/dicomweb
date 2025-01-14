package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/javitab/go-web/router"
	test_suite "github.com/javitab/go-web/tests"
	"github.com/stretchr/testify/assert"
)

// Test Web Router
func TestMain(t *testing.T) {
	tearDown := test_suite.SetupSuite(t)
	defer tearDown(t)
	// Start Web Server
	router := router.AppRouter()

	// Send HTTP Request for Hello World
	expectedResponse :=
		`Web Test Successful`
	req, _ := http.NewRequest("GET", "/web/test", nil)
	req.Header.Set("content-type", "application/json;")
	req.Host = os.Getenv("HTTP_HOST")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	assert.Equal(t, expectedResponse, string(responseData))
	assert.Equal(t, http.StatusOK, w.Code)
}
