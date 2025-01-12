package pgsql

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPgSQLConnection(config Config) (db *gorm.DB, err error) {
	// Define data source name (DSN)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=%s TimeZone=%s",
					   config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode, config.Timezone)

	// Connect using GORM
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}