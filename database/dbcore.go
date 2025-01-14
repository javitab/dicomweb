package database

import (
	"flag"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var IsTestRunning bool

var db_conn *gorm.DB

func MigrateSchemas(db *gorm.DB) {
	// Migrate the schema
	db.AutoMigrate(ServerEvent{})
	db.AutoMigrate(User{})
	db.AutoMigrate(APIKey{})
	db.AutoMigrate(Group{})
	db.AutoMigrate(SecPoint{})

}

func GetDBConn() *gorm.DB {
	if db_conn != nil {
		return db_conn
	}
	if flag.Lookup("test.v") == nil {
		dsn := os.Getenv("DB_DSN")
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		db_conn = db
		return db_conn
	} else {
		return GetTestDBConn()
	}

}

func GetTestDBConn() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}
	return db
}

func InitializeDB() {
	// Get the database connection
	db := GetDBConn()
	MigrateSchemas(db)

}

func InitializeTestDB() {
	// Get Database

	db := GetTestDBConn()
	MigrateSchemas(db)

}
