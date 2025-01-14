package auth

import dbase "github.com/javitab/go-web/database"

type GroupInfo struct {
	DB dbase.Group
}

func GetGroupInfo(GroupID int) GroupInfo {
	// Get Database Connection
	db := dbase.GetDBConn()

	// Get Group from Database
	var GroupInfo GroupInfo
	db.Model(&GroupInfo.DB).Preload("AddSecPoints").Preload("DelSecPoints").Preload("OvrSecPoints").Where("ID = ?", GroupID).Find(&GroupInfo.DB)

	// ### ###
	// ### ###
	// ### ### Assign Functions
	// ### ###
	// ### ###

	return GroupInfo

}
