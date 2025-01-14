package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	auth "github.com/javitab/go-web/auth"
	dbase "github.com/javitab/go-web/database"
	"github.com/javitab/go-web/helpers"
	"golang.org/x/term"
)

var LoggedInUser auth.UserInfo

var MenuFuncs = map[string]func(){
	"LDAP User Login Test":                            CLILDAPAuthenticationTest,
	"Get LDAP User Info":                              CLILDAPGetUserInfo,
	"Standard User Login Test":                        CLIUserLoginTest,
	"Get User Info":                                   CLIGetUserInfo,
	"Change User Password":                            CLIChangePassword,
	"Soft Delete User":                                CLIDeleteUser,
	"Load Groups and Security Points from file":       CLILoadGroupsSecPoints,
	"Add User to Group":                               CLIAddUserToGroup,
	"Remove User From Group":                          CLIRemoveUserFromGroup,
	"Evaluate User Security Points":                   CLIEvalUserSecurity,
	"Update Security Points":                          CLIUserUpdateSecPoints,
	"Set LDAP User":                                   CLISetLDAPUser,
	"Get Group Info":                                  CLIGetGroupInfo,
	"Create User":                                     CLICreateUser,
	"Create API Key":                                  CLICreateAPIKey,
	"Load Groups and Security Points from Embeddings": CLIMigratedGroupsSecPointsEmbedded,
}

func CLIMigratedGroupsSecPointsEmbedded() {
	dbase.CreateSecPoints(nil)
	dbase.CreateGroups(nil)
}

func CLICreateAPIKey() {

	//Get Inputs
	var Desc string
	fmt.Print("Enter key description: ")
	fmt.Scanln(&Desc)
	key, err := dbase.CreateAPIKey(LoggedInUser.DB, Desc)
	if err != nil {
		fmt.Println("Error generating API key")
		return
	}
	helpers.PrettyPrintJSONString(key)
}

func CLICreateUser() {
	CreateNewSuperuser := false

	// Don't check SecPoint if no users exist
	var Users []dbase.User
	db := dbase.GetDBConn()
	db.Model(&dbase.User{}).Find(&Users).Limit(2)
	if len(Users) != 0 {
		// SPCheck
		if sec := LoggedInUser.SPCheck(2); !sec {
			return
		}
	} else {
		fmt.Println("Migrating Groups and SecPoint definitions before first user creation")
		CreateNewSuperuser = true

		// Create Security Points
		dbase.CreateSecPoints(nil)

		// Create Groups
		dbase.CreateGroups(nil)

		fmt.Println("This user will automatically be created as a super user")
	}

	// Get Inputs
	var Username string
	var FirstName string
	var LastName string
	var Email string
	var Password string
	var ConfPassword string
	fmt.Print("Enter Username: ")
	fmt.Scanln(&Username)
	existingUser := auth.GetUserInfo(Username)
	if existingUser.DB.ID != 0 {
		fmt.Printf("User %v already exists", Username)
		return
	}
	fmt.Print("Enter FirstName: ")
	fmt.Scanln(&FirstName)
	fmt.Print("Enter LastName: ")
	fmt.Scanln(&LastName)
	fmt.Print("Enter Email: ")
	fmt.Scanln(&Email)
	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	Password = string(bytePassword)
	fmt.Print("Confirm password: ")
	bytePassword, _ = term.ReadPassword(int(syscall.Stdin))
	ConfPassword = string(bytePassword)
	if ConfPassword != Password {
		fmt.Println("Passwords do not match")
		return
	}

	err := dbase.CreateUser(Username, LastName, FirstName, Email, Password)
	if err != nil {
		fmt.Println("Error while creating user")
		return
	}

	// Get User
	user := auth.GetUserInfo(Username)

	// Set LDAP User
	var action string
	fmt.Print("\nSet IsLdapUser: [y/n]")
	fmt.Scanln(&action)

	// Perform Action
	if action == "y" {
		fmt.Println("Setting IsLdapUser to true")
		user.DB.IsLDAPUser = true
	} else {
		fmt.Println("Setting IsLDAPUser to false")
		user.DB.IsLDAPUser = false
	}

	// Save User
	db.Save(user.DB)

	if CreateNewSuperuser {
		fmt.Printf("Creating Superuser %v", user.DB.Username)
		db.Exec("INSERT INTO user_groups(user_id,group_id) VALUES(? , ?)", user.DB.ID, 1)
	}

}

func CLIGetGroupInfo() {
	// Get Group and confirm exists
	var GroupName string
	fmt.Print("Enter Group Name: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		GroupName = scanner.Text()
		fmt.Printf("\nInput: %q\n", GroupName)
	}
	db := dbase.GetDBConn()
	Group := dbase.Group{}
	Group.Name = GroupName
	db.Model(&Group).Where("name = ?", Group.Name).Find(&Group)
	fmt.Println("Group Input: " + Group.Name)
	if Group.ID == 0 {
		fmt.Printf("Unable to find group %v \n", GroupName)
		return
	}
	GroupInfo := auth.GetGroupInfo(int(Group.ID))
	helpers.PrettyPrintJSONString(GroupInfo)
}

func getSPInput() []auth.SecPointInfo {
	var SPList []auth.SecPointInfo
	for {
		var SPID int
		fmt.Print("Enter SPID: ")
		fmt.Scanln(&SPID)
		SecPoint := auth.GetSecPointInfo(SPID)
		if SecPoint.DB.ID == 0 {
			fmt.Printf("Invalid SPID: %v", SPID)
		} else {
			var action string
			helpers.PrettyPrintJSONString(SecPoint)
			fmt.Print("Add to list? : [y/n] ")
			fmt.Scanln(&action)
			switch action {
			case "y":
				SPList = append(SPList, SecPoint)
			default:
				fmt.Print("SP not added to list")
			}
			var cont string
			fmt.Print("More? : [y/n] ")
			fmt.Scanln(&cont)
			switch cont {
			case "y":
				continue
			default:
				break
			}
		}
		break
	}
	return SPList
}

func CLIUserUpdateSecPoints() {
	// Get User and confirm exists
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)

	user := auth.GetUserInfo(Username)
	if user.DB.ID == 0 {
		fmt.Printf("Unable to find user %v", Username)
		return
	}

	// Get Security Point(s) and confirm exists
	SPList := getSPInput()
	fmt.Printf("%v SP(s) added to list\n", len(SPList))
	for _, sp := range SPList {
		fmt.Printf("   SPID: %v %v\n", sp.DB.ID, sp.DB.Name)
	}
	fmt.Println("")

	// Get Destination for Security Points
	var SPListField string
	fmt.Print("Enter SP Field: [add/del/ovr]")
	fmt.Scanln(&SPListField)

	// Save Security Points to field
	switch SPListField {
	case "add":
		for _, SP := range SPList {
			user.DB.UserAddSecPoints = append(user.DB.UserAddSecPoints, SP.DB)
		}
	case "del":
		for _, SP := range SPList {
			user.DB.UserDelSecPoints = append(user.DB.UserDelSecPoints, SP.DB)
		}
	case "ovr":
		user.DB.UserOvrSecPoints = []dbase.SecPoint{}
		for _, SP := range SPList {
			user.DB.UserOvrSecPoints = append(user.DB.UserOvrSecPoints, SP.DB)
		}
	default:
		fmt.Println("Invalid Selection")
	}

	db := dbase.GetDBConn()
	db.Save(user.DB)

}

func CLISetLDAPUser() {
	// SPCheck
	if sec := LoggedInUser.SPCheck(9); !sec {
		return
	}

	// Get User and confirm exists
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)
	user := auth.GetUserInfo(Username)
	if user.DB.ID == 0 {
		fmt.Printf("Unable to find user %v", Username)
		return
	}

	// Print Current Value
	fmt.Printf("User: %v\nIsLDAPUser Value: %v\n", user.DB.Username, user.DB.IsLDAPUser)

	// Get Action
	var action string
	fmt.Print("\nSet IsLdapUser: [true/false]")
	fmt.Scanln(&action)

	db := dbase.GetDBConn()

	// Perform Action
	if action == "true" {
		fmt.Println("Setting IsLdapUser to true")
		user.DB.IsLDAPUser = true
	} else {
		fmt.Println("Setting IsLDAPUser to false")
		user.DB.IsLDAPUser = false
	}

	db.Save(user.DB)

}

func CLILDAPGetUserInfo() {
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)
	UserInfo, _ := auth.LDAPGetUserInfo(Username)
	data, _ := json.Marshal(UserInfo)
	json_string := string(data)

	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(json_string), "", "	")
	if err != nil {
		dbase.LogServerError("auth_utils:CLILDAPGetUserInfo:jsonSerialize", err, "")
	}
	fmt.Println(prettyJSON.String())

}

func CLILDAPAuthenticationTest() {
	// Get Inputs
	var creds auth.LoginUserInput
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&creds.Username)
	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))

	fmt.Println("")
	creds.Password = string(bytePassword)

	// Validate Credentials

	valid, err := auth.LDAPAuth(creds)
	if err != nil {
		dbase.LogServerError("cli:auth_utils:main:invalid_login", err, "")
		return
	}
	if valid {
		log.Println(
			"Login successful for: " + creds.Username,
		)
		dbase.LogServerEvent("auth_utils:ldap_login", "Login successful for: "+creds.Username, "OK")
	}

	groups := auth.LDAPGroups(creds.Username)
	for idx, group := range groups {
		fmt.Printf("Group %v: %v", idx, group)
	}
}

func CLIUserLoginTest() {
	// Get Inputs
	var creds auth.LoginUserInput
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&creds.Username)
	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))

	fmt.Println("")
	creds.Password = string(bytePassword)

	// Validate Credentials

	token, err := auth.UserLogin(creds, auth.CLILogin, nil)
	if err != nil {
		fmt.Printf("Error validating user credentials: %v \n", err.Error())
		return
	} else {
		fmt.Printf("Login JWT: %v", token)
	}
	fmt.Print("\n")
}

func CLIUserLogin() {

	// Check for CLI API Key
	if os.Getenv("CLI_API_KEY") != "" {
		db := dbase.GetDBConn()
		var APIKey dbase.APIKey
		APIKey.KeyValue = os.Getenv(("CLI_API_KEY"))
		db.Model(dbase.APIKey{}).Find(&APIKey)
		if APIKey.ID == 0 {
			dbase.LogServerEvent("CLIUserLogin:APIKeyNotFound", "API Key not found: "+APIKey.KeyValue, "DENY")
			os.Exit(1)
		}
		var DBUser dbase.User
		DBUser.ID = APIKey.UserID
		db.Model(dbase.User{}).Find(&DBUser)
		LoggedInUser = auth.GetUserInfo(DBUser.Username)
		dbase.LogServerEvent("CLIUserLogin:CLIAPIKeyLogin", fmt.Sprintf("User %v CLI authenticated", DBUser.Username), "AUTH")
		return
	}

	// Get Inputs
	var creds auth.LoginUserInput
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&creds.Username)
	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))

	fmt.Println("")
	creds.Password = string(bytePassword)

	// Validate Credentials
	_, err := auth.UserLogin(creds, auth.CLILogin, nil)
	if err != nil {
		fmt.Printf("Error validating user credentials: %v \n", err.Error())
		panic("user not authenticated")
	} else {
		LoggedInUser = auth.GetUserInfo(creds.Username)
		fmt.Printf("User authenticated: %v", LoggedInUser.DB.Username)
	}
	fmt.Print("\n")
}

func CLIDeleteUser() {
	// Get Inputs
	var DeleteUserID string
	fmt.Print("Enter UserID to (un)delete: ")
	fmt.Scanln(&DeleteUserID)

	// Get User to Delete
	DeleteUserInfo := auth.GetUserInfo(DeleteUserID)
	if DeleteUserInfo.DB.ID == 0 {
		fmt.Println("User does not exist")
		CLIDeleteUser()
	}

	// Display DeleteUserInfo
	helpers.PrettyPrintJSONString(DeleteUserInfo)

	// Get Action
	var GetAction string
	fmt.Println("delete or undelete ? ")
	fmt.Scanln(&GetAction)

	// Get Delete Reason
	var DeleteReason string
	fmt.Print("Enter Reason for (Un)Deletion: ")
	fmt.Scanln(&DeleteReason)

	// Delete user
	DeleteRequest := dbase.DeleteUserRequest{
		Username:       DeleteUserInfo.DB.Username,
		RequestingUser: LoggedInUser.DB.Username,
		Reason:         DeleteReason,
		Action:         GetAction,
	}

	err := dbase.DeleteUser(DeleteRequest)
	if err != nil {
		fmt.Println("Unable to " + string(DeleteRequest.Action) + " user: " + DeleteRequest.Username + "\n" + err.Error())
	}

}

func CLILoadGroupsSecPoints() {

	// Don't check SecPoint if no users exist
	var Users []dbase.User
	db := dbase.GetDBConn()
	db.Model(&dbase.User{}).Find(&Users).Limit(2)
	if len(Users) != 0 {
		// SPCheck
		if sec := LoggedInUser.SPCheck(8); !sec {
			return
		}
	}

	// Get Inputs
	var GroupsYAMLPath string
	fmt.Print("Enter path to groups.yaml [config/auth/groups.yaml]: ")
	fmt.Scanln(&GroupsYAMLPath)
	if GroupsYAMLPath == "" {
		GroupsYAMLPath = "config/auth/groups.yaml"
	}

	var SecPointsYAMLPath string
	fmt.Print("Enter path to secPoints.yaml [config/auth/secPoints.yaml]: ")
	fmt.Scanln(&SecPointsYAMLPath)
	if SecPointsYAMLPath == "" {
		SecPointsYAMLPath = "config/auth/secPoints.yaml"
	}

	// Create Security Points
	dbase.CreateSecPoints(&SecPointsYAMLPath)

	// Create Groups
	dbase.CreateGroups(&GroupsYAMLPath)

}

func CLIAddUserToGroup() {

	db := dbase.GetDBConn()

	// Check Security Point
	if sec := LoggedInUser.SPCheck(7); !sec {
		dbase.LogServerEvent("CLIAddUserToGroup:SPCheck", fmt.Sprintf("user %v does not have sp 7", LoggedInUser.DB.Username), "AUTH")
		return
	}

	// Get Inputs
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)

	// Get User Profile
	User := auth.GetUserInfo(Username)
	if User.DB.ID == 0 {
		panic("User does not exist")
	} else {
		helpers.PrettyPrintJSONString(User)
	}

	// Get Group Name
	fmt.Print("Enter Group Name: ")
	reader := bufio.NewReader(os.Stdin)
	GroupName, _ := reader.ReadString('\n')
	GroupName = strings.TrimSpace(GroupName)

	// Get Group from Database
	Group := dbase.Group{}
	db.Where("name = ?", GroupName).First(&Group)
	if Group.ID == 0 {
		panic("Group does not exist")
	} else {
		helpers.PrettyPrintJSONString(Group)
	}

	var proceed string
	fmt.Print("Add user to group? [y/n]: ")
	fmt.Scanln(&proceed)
	if proceed != "y" {
		return
	}

	// Add user to group
	DBUser := User.DB
	DBUser.Groups = append(DBUser.Groups, Group)
	db.Save(&DBUser)
}

func CLIEvalUserSecurity() {
	// Get Inputs
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)

	// Get User Profile
	User := auth.GetUserInfo(Username)
	if User.DB.ID == 0 {
		panic("User does not exist")
	}

	// Get Security Point to check
	var SPID int
	fmt.Print("Enter Security Point ID: ")
	fmt.Scanln(&SPID)

	// Check if user has security point
	if User.SPCheck(SPID) {
		fmt.Printf("User %v has Security Point %v\n", Username, SPID)
	} else {
		fmt.Printf("User %v does not have Security Point %v\n", Username, SPID)
	}
}

func CLIChangePassword() {
	// Get Inputs
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)

	// Get User Profile
	User := auth.GetUserInfo(Username)
	if User.DB.ID == 0 {
		panic("User does not exist")
	}

	// Get New Password
	fmt.Print("Enter New Password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))

	// Change Password
	err := dbase.ChangeUserPassword(User.DB.Username, string(bytePassword))
	if err != nil {
		fmt.Println("Error changing password: " + err.Error())
	}
}

func CLIGetUserInfo() {
	// Get Inputs
	var Username string
	fmt.Print("Enter Network ID: ")
	fmt.Scanln(&Username)

	UserInfo := auth.GetUserInfo(Username)

	if UserInfo.DB.ID == 0 {
		fmt.Printf("User %v not found\n", Username)
		return
	}

	helpers.PrettyPrintJSONString(UserInfo)
}

func CLIGetSecPointInfo() {
	// Get Inputs
	var SPID int
	fmt.Print("Enter SPID: ")
	fmt.Scanln(&SPID)

	SPInfo := auth.GetSecPointInfo(SPID)

	helpers.PrettyPrintJSONString(SPInfo)
}

func CLIRemoveUserFromGroup() {
	// Get Inputs
	var Username string
	fmt.Print("Enter Username: ")
	fmt.Scanln(&Username)

	// Get User
	User := auth.GetUserInfo(Username)
	if User.DB.ID == 0 {
		fmt.Println("Unable to find user")
	}

	// Get Group
	var GroupName string
	fmt.Print("Enter Group Name: ")
	fmt.Scanln(&GroupName)

	db := dbase.GetDBConn()
	var DBGroup dbase.Group
	db.Model(&DBGroup).Where("name = ?", GroupName).Find(&DBGroup)

	if DBGroup.ID == 0 {
		fmt.Println("Unable to find group")
	}

	err := dbase.RemoveUserFromGroup(User.DB, DBGroup)
	if err != nil {
		fmt.Println("Unable to remove user from group")
		fmt.Println(err.Error())
	}

}
