package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SecPointInfoInput struct {
	SPID int `json:"sec_point_id" binding:"required"`
}

const (
	InvalidInput = "Invalid input"
)

// GetSecPoint godoc
//
//		@Summary		Get information about a given security point
//		@Security		ApiKeyAuth
//		@Schemes		http
//		@Tags			security
//		@Description	Returns information about a single security point including groups with FK relationships
//	 	@Param 			spid	query	int	true	"SPID to search"
//		@Accept			json
//		@Produce		json
//		@Success		200	{object} SecPointInfo
//		@Failure		400	{string}	error message
//		@Router			/auth/sec_point [get]
func GetSecPoint(c *gin.Context) {
	SPID_string := c.Query("spid")
	SPID, err := strconv.Atoi(SPID_string)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/plaintext", []byte("Invalid input"))
		return
	}

	sec_point := GetSecPointInfo(SPID)
	c.JSON(http.StatusOK, sec_point)
}
