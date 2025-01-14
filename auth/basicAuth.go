package auth

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type UserList struct {
	Users map[string]string `yaml:"users"`
}

func loadUserList(filename string) (map[string]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var userList UserList
	err = yaml.Unmarshal(data, &userList)
	if err != nil {
		return nil, err
	}

	return userList.Users, nil
}

func BasicAuth() gin.HandlerFunc {
	users, err := loadUserList("users.yaml")
	if err != nil {
		log.Fatalf("Failed to load user list: %v", err)
	}

	return gin.BasicAuth(users)
}
