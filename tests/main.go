package test_suite

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	dbase "github.com/javitab/go-web/database"
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

func SetupSuite(t *testing.T) func(t *testing.T) {
	log.Println("setup suite")

	// Delete prior copy of DB
	os.Remove("test.db")

	// Initialize DB
	dbase.InitializeTestDB()

	// Get Database Connection
	db := dbase.GetDBConn()

	// Create Security Points
	dbase.CreateSecPoints(nil)
	// Create Groups
	dbase.CreateGroups(nil)

	fmt.Println("This user will automatically be created as a super user")

	// Define User
	err := dbase.CreateUser(
		"testuser",
		"Test",
		"User",
		"testuser@test.com",
		"password",
	)
	if err != nil {
		fmt.Printf("Error occurred during test user creation: %v", err)
	}
	User := dbase.User{
		Username: "testuser",
	}
	db.Model(&User).Find(&User)

	fmt.Printf("Creating Superuser %v", User.Username)
	db.Exec("INSERT INTO user_groups(user_id,group_id) VALUES(? , ?)", User.Username, 1)

	// Get all users
	var users []dbase.User
	db.Model(dbase.User{}).Find(&users)

	assert.Equal(t, len(users), int(1))

	// Return a function to teardown the test
	return func(t *testing.T) {
		log.Println("teardown suite")
	}
}
