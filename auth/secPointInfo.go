package auth

import dbase "github.com/javitab/go-web/database"

type SecPointInfo struct {
	DB                dbase.SecPoint
	ReferencingGroups []GroupInfo
}

func GetSecPointInfo(SPID int) SecPointInfo {
	// Get Database Connection
	db := dbase.GetDBConn()

	// Get SecPoint from Database
	var SecPoint dbase.SecPoint
	db.Model(&SecPoint).Where("id = ?", SPID).Find(&SecPoint)

	// Initialize SecPointInfo w/ DB Object
	SPInfo := SecPointInfo{}
	SPInfo.DB = SecPoint

	// Get groups with reference to SecPoint
	var groupIDs []int
	db.Raw(
		"SELECT id from public.Groups"+
			" INNER JOIN group_add_sec_points"+
			" ON group_add_sec_points.group_id = groups.id"+
			" WHERE group_add_sec_points.sec_point_id = ?",
		SPInfo.DB.ID).Scan(&groupIDs)
	db.Raw(
		"SELECT id from public.Groups"+
			" INNER JOIN group_del_sec_points"+
			" ON group_del_sec_points.group_id = groups.id"+
			" WHERE group_del_sec_points.sec_point_id = ?",
		SPInfo.DB.ID).Scan(&groupIDs)
	db.Raw(
		"SELECT id from public.Groups"+
			" INNER JOIN group_ovr_sec_points"+
			" ON group_ovr_sec_points.group_id = groups.id"+
			" WHERE group_ovr_sec_points.sec_point_id = ?",
		SPInfo.DB.ID).Scan(&groupIDs)
	var GroupInfos []GroupInfo
	for _, id := range groupIDs {
		GroupInfos = append(GroupInfos, GetGroupInfo(id))
	}
	SPInfo.ReferencingGroups = GroupInfos

	// ### ###
	// ### ###
	// ### ### Assign Functions
	// ### ###
	// ### ###

	return SPInfo

}
