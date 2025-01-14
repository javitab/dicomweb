package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ServerRunID string

type ServerEvent struct {
	gorm.Model
	UUID_ID     string
	ServerRunID string
	Archived    bool      `gorm:"default:false"`
	DateTime    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	EventType   string    `gorm:"default:'LoggedEvent'"`
	Details     string
	Status      string
}

func CreateServerStartEvent() {
	db := GetDBConn()
	se := &ServerEvent{}
	se.EventType = "StartingServer"
	se.Details = "Starting Server"
	se.Status = "PENDING"
	se.UUID_ID = uuid.NewString()
	se.ServerRunID = se.UUID_ID
	db.Create(&se)
	ServerRunID = se.ServerRunID
}

func CreateServerStartFailureEvent(
	start_error error,
) {
	err_str := fmt.Sprintf("%v", start_error)
	db := GetDBConn()
	se := &ServerEvent{}
	se.EventType = "StartServerFailure"
	se.Details = err_str
	se.Status = "FAIL"
	se.UUID_ID = uuid.NewString()
	se.ServerRunID = ServerRunID
	db.Create(&se)

}

func LogServerError(
	EventType string,
	err error,
	details string,
) {
	// err_str := fmt.Sprintf("%v", err)
	err_str := err.Error()

	err_str = "Error: \n" + err_str + "\nDetails: \n" + details
	LogServerEvent(EventType, err_str, "ERROR")
	//log.Printf("Server event logged:\nEventType: %v\nError: %v\nDetails: %v\n", EventType, err, details)
}

func LogServerEvent(
	EventType string,
	Details string,
	Status string,
) {

	db := GetDBConn()
	se := &ServerEvent{
		ServerRunID: ServerRunID,
		EventType:   EventType,
		Details:     Details,
		Status:      Status,
		UUID_ID:     uuid.NewString(),
	}
	db.Create(&se)

	fmt.Printf("Server event logged:\n     EventType: %v\n     Details: %v\n     Status: %v\n", EventType, Details, Status)
}
