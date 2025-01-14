package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dbase "github.com/javitab/go-web/database"
)

func ServerEventHandler(c *gin.Context) {
	if c.Request.Method == "GET" {
		GETServerEvents(c)
	} else if c.Request.Method == "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
	}
}

type GetServerEventsResponse struct {
	Limit  int                 `json:"limit"`
	Events []dbase.ServerEvent `json:"events"`
}

// GetServerEvents godoc
//
//		@Summary		Get Logged Server Events
//		@Schemes		http
//		@Tags			api
//		@Security		ApiKeyAuth
//		@Description	Gets all server events from database matching filter criteria
//	 	@Param 			limit query string false "search filter"
//	 	@Param 			EventType query string false "EventType to filter for"
//	 	@Param 			ServerRunID query string false "ServerRunID to filter for"
//		@Accept			json
//		@Produce		json
//		@Success		200	{object} GetServerEventsResponse
//		@Router			/api/server_events [get]
func GETServerEvents(c *gin.Context) {
	// Get the database connection
	db := dbase.GetDBConn()

	// Check if the limit parameter is present
	var limit int
	var err error
	if c.Query("limit") == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid limit parameter",
			})
		}
	}

	//Check if the EventType parameter is present
	var eventType string
	if c.Query("EventType") != "" {
		eventType = c.Query("EventType")
	}

	//Check if the ServerRunID parameter is present
	var ServerRunID string
	if c.Query("ServerRunID") != "" {
		ServerRunID = c.Query("ServerRunID")
	}

	// Get the server events
	var serverEvents []dbase.ServerEvent
	db.Limit(limit).Where(&dbase.ServerEvent{
		EventType:   eventType,
		ServerRunID: ServerRunID,
	}).Find(&serverEvents)

	// Return the server events as JSON
	c.JSON(http.StatusOK, gin.H{
		"limit":  limit,
		"events": serverEvents})
}
