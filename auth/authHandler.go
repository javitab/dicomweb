package auth

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	dbase "github.com/javitab/go-web/database"
)

type CreateUserInput struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
}

type CreateUserResponse struct {
	Message string `json:"message" enums:"User created,User already exists,Error creating user"`
}

// CreateUser godoc
//
//		@Summary		Create a new user via API
//		@Security		ApiKeyAuth
//		@Schemes		http
//		@Tags			user/group security
//		@Description	Create a new user given a CreateUserInput object
//	 	@Param request body CreateUserInput true "query params"
//		@Accept			json
//		@Produce		json
//		@Success		200	{object} CreateUserResponse
//		@Router			/auth/create_user [post]
func CreateUser(c *gin.Context) {
	reqUser, _ := c.Get("user")
	LoggedInUser := reqUser.(UserInfo)
	if !LoggedInUser.SPCheck(2) {
		err := fmt.Errorf("user %v missing security point 2", LoggedInUser.DB.Username)
		dbase.LogServerError("CreateUser:HTTP:MissingSecurityPoint", err, "AUTH")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"err":   fmt.Sprintf("%v", err),
		})
		return
	}
	var input CreateUserInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
			"err":   fmt.Sprintf("%v", err),
			"input": input,
		})
		dbase.LogServerError("CreateUser:HTTP:InvalidInput", err, "Invalid Input for CreateUser")
		return
	}
	db := dbase.GetDBConn()

	//Check if user exists
	var userFound dbase.User
	db.Where("username = ?", input.Username).Find(&userFound)

	if userFound.Username != "" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
		return
	}

	//Create user
	new_user_err := dbase.CreateUser(input.Username, input.LastName, input.FirstName, input.Email, input.Password)
	if new_user_err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating user",
		})
		dbase.LogServerError("CreateUser:HTTP", new_user_err, "Error creating user: "+input.Username)
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"message": "User created",
		})
	}
}

type LoginUserInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUserResponse struct {
	Token string `json:"token"`
}

// LoginUser godoc
//
//		@Summary		Login user
//		@Schemes		http
//		@Tags			login
//		@Description	Authenticate a user given a LoginUserInput object. Returns a JWT token upon successful authentication
//	 	@Param request body LoginUserInput true "query params"
//		@Accept			json
//		@Produce		plain
//		@Success		200	{string}	token
//		@Router			/auth/login [post]
func LoginUser(c *gin.Context) {
	var input LoginUserInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
			"err":   fmt.Sprintf("%v", err),
		})
		dbase.LogServerError("LoginUser:HTTP:InvalidInput", err, "Invalid Input for LoginUser")
		return
	}
	token, err := UserLogin(input, WebLogin, c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unable to authenticate",
			"err":   fmt.Sprintf("%v", err),
		})
		dbase.LogServerError("LoginUser:HTTP:InvalidLogin", err, "Unable to authenticate")
		return
	}
	token = "Bearer " + token
	c.Data(http.StatusOK, "text/plaintext", []byte(token))
	c.SetCookie("token", token, int(time.Hour*1), "/", "localhost", false, true)
}

type GetJWTFromAPIKeyInput struct {
	Key string `json:"key" binding:"required"`
}

// GetJWTFromAPIKey godoc
//
//		@Summary		Get JWT from API Key
//		@Schemes		http
//		@Tags			login
//		@Description	Authenticate a user given an APIKey. Returns a JWT token upon successful authentication
//	 	@Param request body GetJWTFromAPIKeyInput true "query params"
//		@Accept			json
//		@Produce		plain
//		@Success		200	{string}	token
//		@Router			/auth/generate_jwt [post]
func GetJWTFromAPIKey(c *gin.Context) {
	var api_key_input GetJWTFromAPIKeyInput
	err := c.ShouldBindJSON(&api_key_input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
			"err":   fmt.Sprintf("%v", err),
			"input": api_key_input,
		})
		dbase.LogServerError("GetJWT:HTTP:InvalidInput", err, "Invalid input")
		return
	}
	db := dbase.GetDBConn()
	var APIKey dbase.APIKey
	APIKey.KeyValue = api_key_input.Key
	db.Model(dbase.APIKey{}).Find(&APIKey)
	if APIKey.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API Key not found",
		})
		dbase.LogServerEvent("GetJWT:HTTP:APIKeyNotFound", "API Key not found: "+api_key_input.Key, "ERROR")
		return
	}
	var user dbase.User
	db.Where("ID = ?", APIKey.UserID).Find(&user)
	token, err := dbase.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generating token",
		})
		dbase.LogServerError("GetJWT:HTTP", err, "Error generating token for user: "+user.Username)
		return
	}
	token = "Bearer " + token
	c.Data(http.StatusOK, "text/plaintext", []byte(token))
}

type GetUserInput struct {
	Username string `json:"username" binding:"required"`
}

// GetUser godoc
//
//		@Summary		Get user info
//		@Security		ApiKeyAuth
//		@Schemes		http
//		@Tags			user/group security
//		@Description	Given a username, gives details about user
//	 	@Param 			username query string true "username to lookup"
//		@Accept			json
//		@Produce		json
//		@Success		200	{object} UserInfo
//		@Router			/auth/user [get]
func GetUser(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
			"input": username,
		})
		log.Print("Username not provided")
		return
	}
	UserInfo := GetUserInfo(username)
	c.JSON(http.StatusOK, UserInfo)
}

type GenerateAPIKeyResponse struct {
	User    string `json:"user"`
	Message string `json:"message" enums:"API Key generated"`
	APIKey  string
}

// GenerateAPIKey godoc
//
//		@Summary		Generate API Key
//		@Security		ApiKeyAuth
//		@Schemes		http
//		@Tags			user/group security
//		@Description	Given a username, gives details about user
//	 	@Param description	formData	string	true	"desc for key usage"
//		@Accept			json
//		@Produce		json
//		@Success		200	{object} GenerateAPIKeyResponse
//		@Router			/auth/generate_api_key [post]
func GenerateAPIKey(c *gin.Context) {
	// Get user from context
	db := dbase.GetDBConn()
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	var userData dbase.User
	db.Where("Username = ?", user).Find(&userData)
	apiKey, err := dbase.CreateAPIKey(userData, c.Query("description"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error generating API Key",
		})
		dbase.LogServerError("GenerateAPIKey:HTTP", err, "Error generating API Key for user: "+userData.Username)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":    userData.Username,
		"message": "API Key generated",
		"api_key": apiKey,
	})
}
