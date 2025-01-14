package auth

import (
	"fmt"
	"sort"

	dbase "github.com/javitab/go-web/database"
)

type UserInfo struct {
	DB             dbase.User
	LDAPGroups     []string
	IsLDAPUser     bool
	IsActiveUser   bool
	SecurityPoints map[uint]EvalSP

	// Lookups (add tag to remove from JSON)
	SPCheck func(SPID int) bool `json:"-"`

	// Functions (add tag to remove from JSON)
	SetLDAPUser        func(bool) error                   `json:"-"`
	SetUserSecPoint    func(SPID int, field string) error `json:"-"`
	AddUserToGroup     func(GID int) error                `json:"-"`
	RemoveUserSecPoint func(SPID int, field string) error `json:"-"`
	GenerateAPIKey     func(desc string) error            `json:"-"`
}

func GetUserInfo(Username string) UserInfo {

	// Get Database Connection
	db := dbase.GetDBConn()

	// Get User from Database
	var DBUser dbase.User
	db.
		Unscoped().
		Preload("UserAddSecPoints").
		Preload("UserDelSecPoints").
		Preload("UserOvrSecPoints").
		Preload("APIKeys").
		Preload("Groups").
		Preload("Groups.AddSecPoints").
		Preload("Groups.DelSecPoints").
		Preload("Groups.OvrSecPoints").
		Where("Username = ?", Username).Find(&DBUser)

	// Populate UserInfo Object with DB values
	var UserInfo UserInfo
	UserInfo.DB = DBUser

	// Populate attributes
	UserInfo.IsActiveUser = !UserInfo.DB.DeletedAt.Valid
	UserInfo.SecurityPoints = enumSecurityPoints(UserInfo.DB)

	// LDAP Attributes
	UserInfo.IsLDAPUser = UserInfo.DB.IsLDAPUser
	if UserInfo.IsLDAPUser {
		UserInfo.LDAPGroups = LDAPGroups(Username)
	}

	// ### ###
	// ### ###
	// ### ### Assign Functions
	// ### ###
	// ### ###

	// ###
	// ### SPCheck Function
	// ###

	UserInfo.SPCheck = func(SPID int) bool {
		_, exists := UserInfo.SecurityPoints[uint(SPID)]
		if !exists {
			if _, exists := UserInfo.SecurityPoints[uint(1)]; exists {
				dbase.LogServerEvent("SPcheck", fmt.Sprintf("User: %v SPID: %v", UserInfo.DB.Username, SPID), "SUPERUSER")
				return true
			}
			fmt.Printf("SPCheck %v - DENIED\n", SPID)
			dbase.LogServerEvent("SPcheck", fmt.Sprintf("User: %v SPID: %v", UserInfo.DB.Username, SPID), "DENY")
		}
		return exists
	}

	// ###
	// ### SetUserSecPoint Function
	// ###

	UserInfo.SetUserSecPoint = func(SPID int, field string) error {
		// Check if user already has SecPoint
		if UserInfo.SPCheck(SPID) {
			return fmt.Errorf("user already has security point: %v\nsource: %v", SPID, UserInfo.SecurityPoints[uint(SPID)].Source)
		}

		SetSecPoint := GetSecPointInfo(SPID)
		switch field {
		case "UserAddSecPoints":
			UserInfo.DB.UserAddSecPoints = append(UserInfo.DB.UserAddSecPoints, SetSecPoint.DB)
		case "UserDelSecPoints":
			UserInfo.DB.UserDelSecPoints = append(UserInfo.DB.UserDelSecPoints, SetSecPoint.DB)
		case "UserOvrSecPoints":
			UserInfo.DB.UserDelSecPoints = append(UserInfo.DB.UserDelSecPoints, SetSecPoint.DB)
		default:
			return fmt.Errorf("invalid field selection: %q", field)
		}
		db := dbase.GetDBConn()
		db.Save(UserInfo.DB)

		return nil
	}

	// ###
	// ### RemoveUserSecPoint Function
	// ###

	UserInfo.RemoveUserSecPoint = func(SPID int, field string) error {
		SetSecPoint := GetSecPointInfo(SPID)
		switch field {
		case "UserAddSecPoints":
			db.Exec("DELETE FROM user_add_sec_points where user_id=? and sec_point_id=?", UserInfo.DB.ID, SetSecPoint.DB.ID)
		case "UserDelSecPoints":
			db.Exec("DELETE FROM user_del_sec_points where user_id=? and sec_point_id=?", UserInfo.DB.ID, SetSecPoint.DB.ID)
		case "UserOvrSecPoints":
			db.Exec("DELETE FROM user_ovr_sec_points where user_id=? and sec_point_id=?", UserInfo.DB.ID, SetSecPoint.DB.ID)
		default:
			return fmt.Errorf("invalid field selection: %q", field)
		}
		UserInfo = GetUserInfo(UserInfo.DB.Username)
		return nil
	}

	// ###
	// ### AddUserToGroup Function
	// ###

	UserInfo.AddUserToGroup = func(GID int) error {
		group := GetGroupInfo(GID)
		// Check if user in group
		UserInGroup := false
		for _, g := range UserInfo.DB.Groups {
			if g.ID == group.DB.ID {
				// GroupIndex = indx
				UserInGroup = true
				break
			}
		}
		if UserInGroup {
			return fmt.Errorf("user already in group")
		}

		// Add user to group
		db := dbase.GetDBConn()
		db.Model(&UserInfo.DB).Association("Groups").Append(&group.DB)

		// Log event
		dbase.LogServerEvent("UpdateUser:HTTP", "User added to group: "+UserInfo.DB.Username+"\nGroup: "+group.DB.Name, "INFO")
		return nil
	}

	// ###
	// ### GenerateAPIKey Function
	// ###

	UserInfo.GenerateAPIKey = func(desc string) error {
		// Generate API Key
		APIKey, err := dbase.CreateAPIKey(UserInfo.DB, desc)
		if APIKey == nil {
			dbase.LogServerError("UserInfo:GenerateAPIKey", err, "Error generating API Key for user: "+UserInfo.DB.Username)
			return fmt.Errorf("error generating API Key")
		}
		return nil
	}

	// ### ###
	// ### ### Return Object
	// ### ###

	return UserInfo
}

type EvalSP struct {
	Group  *dbase.Group `json:"-"`
	Source string       `json:"source"`
	SP     dbase.SecPoint
}

func enumSecurityPoints(user dbase.User) (SecDict map[uint]EvalSP) {
	var SecPoints []EvalSP

	// ### ###
	// ### ### Evaluate Group Security Points
	// ### ###

	// Create a list of all User Groups ordered by Priority
	PrioritizedUserGroups := user.Groups
	sort.Slice(PrioritizedUserGroups, func(i, j int) bool {
		return PrioritizedUserGroups[i].Priority < PrioritizedUserGroups[j].Priority
	})

	// Enumerate Group secPoint fields
	for _, group := range PrioritizedUserGroups {

		// Evaluate Add Security Points and add to SecPoints if missing
		for _, sp := range group.AddSecPoints {
			// Check if SP already exists in SecPoints
			var exists bool
			for _, spEval := range SecPoints {
				if spEval.SP.ID == sp.ID {
					exists = true
					break
				}
			}
			// If SP does not exist, add it
			if !exists {
				SecPoints = append(SecPoints, EvalSP{
					Group:  &group,
					Source: fmt.Sprintf("%v", group.Name) + ":AddSecPoints",
					SP:     sp,
				})
			}
		}

		// Evaluate Delete Security Points and remove from SecPoints if present
		for _, sp := range group.DelSecPoints {
			// Check if SP exists in SecPoints
			for i, spEval := range SecPoints {
				if spEval.SP.ID == sp.ID {
					// Remove SP from SecPoints
					SecPoints = append(SecPoints[:i], SecPoints[i+1:]...)
					break
				}
			}
		}

	}

	// Take last item from PrioritizedUserGroups and evaluate Override Security Points
	if len(PrioritizedUserGroups) > 0 {
		group := PrioritizedUserGroups[len(PrioritizedUserGroups)-1]
		var GroupOvrSecPoints []EvalSP
		for _, sp := range group.OvrSecPoints {
			GroupOvrSecPoints = append(GroupOvrSecPoints, EvalSP{
				Group:  &group,
				Source: fmt.Sprintf("%v", group.Name) + ":OvrSecPoints",
				SP:     sp,
			})
		}
		if len(GroupOvrSecPoints) > 0 {
			SecPoints = GroupOvrSecPoints
		}
	}

	// ### ###
	// ### ### Evaluate User Security Points
	// ### ###

	// Enumerate User secPoint fields
	var UserAddSecPoints []EvalSP
	for _, sp := range user.UserAddSecPoints {
		UserAddSecPoints = append(UserAddSecPoints, EvalSP{
			Group:  nil,
			Source: "User" + ":AddSecPoints",
			SP:     sp,
		})
	}
	var UserDelSecPoints []EvalSP
	for _, sp := range user.UserDelSecPoints {
		UserDelSecPoints = append(UserDelSecPoints, EvalSP{
			Group:  nil,
			Source: "User" + ":DelSecPoints",
			SP:     sp,
		})
	}
	var UserOvrSecPoints []EvalSP
	for _, sp := range user.UserOvrSecPoints {
		UserOvrSecPoints = append(UserOvrSecPoints, EvalSP{
			Group:  nil,
			Source: "User" + ":OvrSecPoints",
			SP:     sp,
		})
	}

	// Evaluate User Add Security Points and add to SecPoints if missing
	for _, sp := range UserAddSecPoints {
		// Check if SP already exists in SecPoints
		var exists bool
		for _, spEval := range SecPoints {
			if spEval.SP.ID == sp.SP.ID {
				exists = true
				break
			}
		}
		// If SP does not exist, add it
		if !exists {
			SecPoints = append(SecPoints, sp)
		}
	}

	// Evaluate User Delete Security Points and remove from SecPoints if present
	for _, sp := range UserDelSecPoints {
		// Check if SP exists in SecPoints
		for i, spEval := range SecPoints {
			if spEval.SP.ID == sp.SP.ID {
				// Remove SP from SecPoints
				SecPoints = append(SecPoints[:i], SecPoints[i+1:]...)
				break
			}
		}
	}

	// Evaluate User Override Security Points
	if len(UserOvrSecPoints) > 0 {
		SecPoints = UserOvrSecPoints
	}

	// Enumerate SecPoints to SecDict
	SecDict = make(map[uint]EvalSP)
	for _, sp := range SecPoints {
		SecDict[sp.SP.ID] = sp
	}

	return SecDict
}
