package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetGroup godoc
//
//		@Summary		Get group info
//		@Security		ApiKeyAuth
//		@Schemes		http
//		@Tags			security
//		@Description	Given a group id, gives details about group
//	 	@Param 			group_id query int true "groupid to lookup"
//		@Accept			json
//		@Produce		json
//		@Success		200	{object} 	UserInfo
//		@Failure		400	{string}	error message
//		@Router			/auth/group [get]
func GetGroup(c *gin.Context) {
	var groupID int
	groupID, err := strconv.Atoi(c.Query("group_id"))
	if err != nil {
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Invalid input"))
	}

	group := GetGroupInfo(groupID)
	c.JSON(http.StatusOK, gin.H{"group": group})
}
