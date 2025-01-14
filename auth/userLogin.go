package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	dbase "github.com/javitab/go-web/database"
	"golang.org/x/crypto/bcrypt"
)

type ValidLoginMode string

const (
	WebLogin ValidLoginMode = "web_login"
	CLILogin ValidLoginMode = "cli_login"
)

func UserLogin(input LoginUserInput, login_mode ValidLoginMode, c *gin.Context) (string, error) {

	//Check if user exists
	userFound := GetUserInfo(input.Username)

	if userFound.DB.Username == "" {
		dbase.LogServerEvent("UserLogin:UserNotFound", "User not found: "+input.Username, "LOGIN")
		return "", fmt.Errorf(("user not found"))
	}

	if userFound.DB.DeletedAt.Valid {
		dbase.LogServerEvent("UserLogin:UserDeleted", "User deleted: "+input.Username, "LOGIN")
		return "", fmt.Errorf("user deleted")
	}

	if userFound.DB.IsLDAPUser {
		// Check if user in LDAP
		if exists := LDAPUserExists(userFound.DB.Username); exists {
			// Check password against LDAP if user exists
			authenticated, err := LDAPAuth(input)
			if (!authenticated) && (err != nil) {
				dbase.LogServerEvent("UserLogin:InvalidLDAPPassword", fmt.Sprintf("Invalid password for user: %v", input.Username), "LOGIN")
				return "", fmt.Errorf("invalid ldap password")
			}
			LDAPEvalGroups(userFound)
		}
	} else {
		//Check password against database if user does not exist
		if err := bcrypt.CompareHashAndPassword([]byte(userFound.DB.Password), []byte(input.Password)); err != nil {
			dbase.LogServerEvent("UserLogin:InvalidPassword", "Invalid password for user: "+input.Username, "LOGIN")
			return "", fmt.Errorf("invalid user password")
		}
	}

	// Check if user has login permissions

	switch login_mode {
	case CLILogin:
		if !userFound.SPCheck(4) {
			dbase.LogServerEvent("UserLogin:Unauthorized", "User not authorized for CLI login: "+input.Username, "LOGIN")
			return "", fmt.Errorf("unauthorized login: missing Security Point 4")
		}
	case WebLogin:
		if !userFound.SPCheck(5) {
			dbase.LogServerEvent("UserLogin:Unauthorized", "User not authorized for Web login: "+input.Username, "LOGIN")
			return "", fmt.Errorf("unauthorized login: missing Security Point 5")

		}
	}

	//
	// Generate JWT
	token, err := dbase.GenerateJWT(userFound.DB.Username)
	if err != nil {
		dbase.LogServerError("UserLogin:ErrorGeneratingToken", err, "Error generating token for user: "+input.Username)
		return "", fmt.Errorf("unable to generate token")
	}

	dbase.LogServerEvent("UserLogin:UserLoggedIn", "User logged in: "+input.Username, "LOGIN")
	return token, nil
}
