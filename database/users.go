package database

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/javitab/go-web/helpers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UUID_ID          uuid.UUID
	Username         string `json:"username" gorm:"unique"`
	LastName         string
	FirstName        string
	Email            string     `json:"email" gorm:"unique"`
	Password         string     `json:"-"`
	IsLDAPUser       bool       `default:"false"`
	APIKeys          []APIKey   `gorm:"foreignKey:UserID" json:"-"`
	Groups           []Group    `gorm:"many2many:user_groups;"`
	UserAddSecPoints []SecPoint `gorm:"many2many:user_add_sec_points;"`
	UserDelSecPoints []SecPoint `gorm:"many2many:user_del_sec_points;"`
	UserOvrSecPoints []SecPoint `gorm:"many2many:user_ovr_sec_points;"`
}

type APIKey struct {
	gorm.Model
	UserID         uint
	UUID_ID        uuid.UUID
	expirationDate time.Time
	KeyValue       string `gorm:"unique"`
	Description    string
}

func GenerateJWT(username string) (string, error) {

	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})
	token, err := generateToken.SignedString([]byte(os.Getenv("SECRET_JWT_KEY")))
	if err != nil {
		return "", err
	}
	return token, nil
}

func CreateUser(
	Username string,
	LastName string,
	FirstName string,
	Email string,
	Password string,
) error {
	db := GetDBConn()

	//Check if user exists
	user := db.Where("username = ?", Username).Find(User{})
	if user.RowsAffected > 0 {
		return errors.New("User already exists")
	}

	//Generate password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		LogServerError("CreateUser:GeneratePasswordHash", err, "Error generating password hash for user: "+Username)
		return errors.New("error creating user")
	}

	//Create new user
	newUser := &User{
		Username:  Username,
		LastName:  LastName,
		FirstName: FirstName,
		Email:     Email,
		Password:  string(passwordHash),
		UUID_ID:   uuid.New(),
	}

	//Save to database
	db.Create(&newUser)

	//Log Server Event
	LogServerEvent("CreateUser", "User created", "INFO")

	return nil
}

func ChangeUserPassword(Username string, NewPassword string) error {
	db := GetDBConn()

	//Check if user exists
	var user User
	db.Where("username = ?", Username).Find(&user)
	if user.ID == 0 {
		return errors.New("user not found")
	}

	//Generate password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(NewPassword), bcrypt.DefaultCost)
	if err != nil {
		LogServerError("ChangeUserPassword:GeneratePasswordHash", err, "Error generating password hash for user: "+Username)
		return errors.New("error changing password")
	}

	user.Password = string(passwordHash)

	_ = helpers.PrettyPrintJSONString(user)

	db.Save(&user)

	//Log Server Event
	LogServerEvent("ChangeUserPassword", "Password changed", "INFO")

	return nil
}

type DeleteUserRequest struct {
	Username       string
	RequestingUser string
	Reason         string
	Action         string
}

func DeleteUser(input DeleteUserRequest) error {
	db := GetDBConn()
	var DeleteUser User
	//Check if user exists
	db.Unscoped().Where("username = ?", input.Username).Find(&DeleteUser)
	if DeleteUser.ID == 0 {
		del_err := errors.New("user for delete not found")
		LogServerError("DeleteUser:CheckForUser", del_err, "Unable to find user for deletion: "+input.Username)
		return del_err
	}

	switch input.Action {
	case "delete":
		if DeleteUser.DeletedAt.Valid {
			return errors.New("user already deleted, cannot delete")
		}
		db.Delete(&DeleteUser)
		LogServerEvent("DeleteUser", fmt.Sprintf("User deleted: %v\nDeleted by: %v\nReason: %v", input.Username, input.RequestingUser, input.Reason), "INFO")
	case "undelete":
		if !DeleteUser.DeletedAt.Valid {
			return errors.New("user currently active, cannot undelete")
		}
		db.Unscoped().Model(&User{}).Where("id", DeleteUser.ID).Update("deleted_at", nil)
		LogServerEvent("DeleteUser", fmt.Sprintf("User undeleted: %v\nUndeleted by: %v\nReason: %v", input.Username, input.RequestingUser, input.Reason), "INFO")
	default:
		fmt.Println("Invalid option selected, please choose delete or undelete")
	}

	return nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func CreateAPIKey(u User, description string) (*APIKey, error) {
	db := GetDBConn()

	//Generate Secure String for API Key
	apiKey, err := GenerateRandomString(64)
	if err != nil {
		LogServerError("CreateAPIKey:GenerateKey", err, "Error generating API Key for user: "+u.Username)
		return nil, errors.New("error generating API Key")
	}

	//Create new API Key
	newAPIKey := &APIKey{
		UserID:         u.ID,
		expirationDate: time.Now().AddDate(0, 1, 0),
		KeyValue:       apiKey,
		UUID_ID:        uuid.New(),
		Description:    description,
	}

	//Save to database
	db.Create(&newAPIKey)

	return newAPIKey, nil
}

func AddUserToGroup(u User, g Group) error {
	db := GetDBConn()

	//Check if user is already in group
	if db.Model(&u).Association("Groups").Find(&g); g.ID != 0 {
		return errors.New("user already in group")
	}

	//Add user to group
	db.Model(&u).Association("Groups").Append(&g)

	return nil
}

func RemoveUserFromGroup(u User, g Group) error {
	db := GetDBConn()

	//Check if user is already in group
	if db.Model(&u).Association("Groups").Find(&g); g.ID == 0 {
		return errors.New("user not in group")
	}

	//Remove from Group
	db.Model(&u).Association("Groups").Delete([]Group{g})

	return nil
}
