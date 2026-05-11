package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// ConnectDB thiết lập kết nối database với GORM
func GormConnectDB(dbType string, connectionString string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch dbType {
	case "mysql":
		dialector = mysql.Open(connectionString)
	case "postgres":
		dialector = postgres.Open(connectionString)
	case "sqlite":
		dialector = sqlite.Open(connectionString)
	case "sqlserver":
		dialector = sqlserver.Open(connectionString)
	default:
		return nil, fmt.Errorf("database type not supported: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Printf("Database connection error: %v", err)
		return nil, err
	}

	log.Println("Database connection successful!")
	return db, nil
}
