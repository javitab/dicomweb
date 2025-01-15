package auth

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dbase "github.com/javitab/go-web/database"
)

type UserAction struct {
	Username string `json:"username" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
	Action   string `json:"action" binding:"required"`
	Value    string `json:"value"`
}

// UpdateUser godoc
//
//		@Summary		Update User Record
//		@Security		ApiKeyAuth
//		@Schemes		http
//		@Tags			user/group security
//		@Description	Given a username, will make given updates
//	 	@Param 			username query string true "username to update"
//	 	@Param 			action query string true "action to perform" Enums(delete_user,undelete_user,add_group,remove_group,add_user_sec_point,remove_user_sec_point)
//	 	@Param 			reason query string true "reason for update (incident #, etc.)"
//	 	@Param 			value query string true "value to set"
//	 	@Param 			sec_point_field query string false "field to append user-level security point to" Enums(UserAddSecPoints,UserDelSecPoints,UserOvrSecPoints)
//		@Accept			json
//		@Produce		plain
//		@Success		200	{string}	operation outcome
//		@Router			/auth/update_user [post]
func UpdateUser(c *gin.Context) {
	reqUser := GetUserInfo(c.GetString("currentUser"))

	// Validate username input
	username := c.Query("username")
	if username == "" {
		err := fmt.Errorf("username not provided")
		dbase.LogServerError("UpdateUser:HTTP:InvalidInput", err, "Invalid Input for UpdateUser")
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Input error: "+err.Error()))
		return
	}

	// Validate action input
	action := c.Query("action")
	if action == "" {
		err := fmt.Errorf("action not provided")
		dbase.LogServerError("UpdateUser:HTTP:InvalidInput", err, "Invalid Input for UpdateUser")
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Input error: "+err.Error()))
		return
	}

	// Validate value input
	value := c.Query("value")
	if value == "" && action != "delete_user" && action != "undelete_user" {
		err := fmt.Errorf("value not provided")
		dbase.LogServerError("UpdateUser:HTTP:InvalidInput", err, "Invalid Input for UpdateUser")
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Input error: "+err.Error()))
		return
	}

	// Validate reason input
	reason := c.Query("reason")
	if reason == "" {
		err := fmt.Errorf("reason not provided")
		dbase.LogServerError("UpdateUser:HTTP:InvalidInput", err, "Invalid Input for UpdateUser")
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Input error: "+err.Error()))
		return
	}

	// Check if user exists
	UserInfo := GetUserInfo(username)
	if UserInfo.DB.ID == 0 {
		err := fmt.Errorf("user %v does not exist", username)
		dbase.LogServerError("UpdateUser:HTTP:InvalidInput", err, "Invalid Input for UpdateUser")
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Input error: "+err.Error()))
		return
	}

	// Update user
	switch action {
	case "delete_user":
		DeleteUserRequest := dbase.DeleteUserRequest{
			Username:       username,
			RequestingUser: reqUser.DB.Username,
			Reason:         reason,
			Action:         action,
		}
		// Don't allow user to delete self
		if username == reqUser.DB.Username {
			err := fmt.Errorf("user cannot delete self")
			dbase.LogServerError("UpdateUser:HTTP:DeleteUser", err, fmt.Sprintf("User %v attempted to delete self", reqUser.DB.Username))
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}
		err := dbase.DeleteUser(DeleteUserRequest)
		if err != nil {
			dbase.LogServerError("UpdateUser:HTTP:DeleteUser", err, "Unable to delete user")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}
	case "undelete_user":
		DeleteUserRequest := dbase.DeleteUserRequest{
			Username:       username,
			RequestingUser: reqUser.DB.Username,
			Reason:         reason,
			Action:         action,
		}
		err := dbase.DeleteUser(DeleteUserRequest)
		if err != nil {
			dbase.LogServerError("UpdateUser:HTTP:UnDeleteUser", err, "Unable to undelete user")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}
	case "remove_group":
		// Convert string to int
		var groupID int
		groupID, _ = strconv.Atoi(value)

		// Check if group exists
		group := GetGroupInfo(groupID)
		if group.DB.ID == 0 {
			err := fmt.Errorf("group not found")
			dbase.LogServerError("UpdateUser:HTTP:RemoveGroup", err, "Group not found")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		// Check if user in group
		UserInGroup := false
		// var GroupIndex int
		for _, g := range UserInfo.DB.Groups {
			if g.ID == group.DB.ID {
				// GroupIndex = indx
				UserInGroup = true
				break
			}
		}
		if !UserInGroup {
			err := fmt.Errorf("user not in group")
			dbase.LogServerError("UpdateUser:HTTP:RemoveGroup", err, "User not in group")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		// Remove user from group
		db := dbase.GetDBConn()
		db.Model(&UserInfo.DB).Association("Groups").Delete(&group.DB)

		// Log event
		dbase.LogServerEvent("UpdateUser:HTTP", "User removed from group: "+username+"\nGroup: "+group.DB.Name, "INFO")
	case "add_group":
		// Convert string to int
		var groupID int
		groupID, _ = strconv.Atoi(value)

		// Check if group exists
		group := GetGroupInfo(groupID)
		if group.DB.ID == 0 {
			err := fmt.Errorf("group not found")
			dbase.LogServerError("UpdateUser:HTTP:AddGroup", err, "Group not found")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		err := UserInfo.AddUserToGroup(int(group.DB.ID))
		if err != nil {
			dbase.LogServerError("UpdateUser:HTTP:AddGroup", err, "reqUser: "+reqUser.DB.Username)
		}

		// Log event
		dbase.LogServerEvent("UpdateUser:HTTP", "User added to group: "+username+"\nGroup: "+group.DB.Name, "INFO")

	case "add_user_sec_point":
		// Validate field input
		field := c.Query("sec_point_field")

		// Convert value to integer
		SPID, err := strconv.Atoi(value)
		if err != nil {
			err := fmt.Errorf("unable to convert value to integer: %q", value)
			dbase.LogServerError("UpdateUser:HTTP:UpdateSecPoints", err, "Invalid SPID value")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		// Perform SPCheck
		if !reqUser.SPCheck(8) {
			err := fmt.Errorf("user %v missing security point 8", reqUser.DB.Username)
			c.Data(http.StatusUnauthorized, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		// Update User Security Points
		err = UserInfo.SetUserSecPoint(SPID, field)
		if err != nil {
			c.Data(http.StatusInternalServerError, "text/plaintext", []byte("error: "+err.Error()))
			return
		}
	case "remove_user_sec_point":
		// Validate field input
		field := c.Query("sec_point_field")

		// Convert value to integer
		SPID, err := strconv.Atoi(value)
		if err != nil {
			err := fmt.Errorf("unable to convert value to integer: %q", value)
			dbase.LogServerError("UpdateUser:HTTP:UpdateSecPoints", err, "Invalid SPID value")
			c.Data(http.StatusBadRequest, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		// Perform SPCheck
		if !reqUser.SPCheck(8) {
			err := fmt.Errorf("user %v missing security point 8", reqUser.DB.Username)
			c.Data(http.StatusUnauthorized, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

		// Update User Security Points
		err = UserInfo.RemoveUserSecPoint(SPID, field)
		if err != nil {
			c.Data(http.StatusInternalServerError, "text/plaintext", []byte("error: "+err.Error()))
			return
		}

	default:
		err := fmt.Errorf("undefined action: %q", action)
		dbase.LogServerError("UpdateUser:HTTP:InvalidInput", err, "Invalid Input for UpdateUser")
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Input error: "+err.Error()))
		return
	}

	c.Data(http.StatusOK, "text/plaintext", []byte("User updated"))
	dbase.LogServerEvent("UpdateUser:HTTP", "User updated: "+username+"\nAction: "+action, "INFO")
}
